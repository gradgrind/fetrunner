package autotimetable

import (
	"fetrunner/internal/base"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"runtime/debug"
	"slices"
	"strings"
	"time"
)

/*
	Default parameter values for autotimetable.

There is probably no general "optimum" value for some parameters, it is
likely to depend on the data. But perhaps values can be found which
are frequently useful. It might be helpful to use shorter overall timeouts
during the initial cycles of testing the data, to identify potential problem
areas without long processing delays. For later cycles longer times may be
necessary (depending on the difficulty of the data).
*/
func DefaultParameters() *Parameters {
	return &Parameters{
		MAXPROCESSES:             MaxProcesses(0),
		TIMEOUT:                  300, // seconds
		NEW_BASE_TIMEOUT_FACTOR:  12,  // => 1.2
		NEW_PHASE_TIMEOUT_FACTOR: 15,  // => 1.5
		LAST_TIME_0:              5,
		LAST_TIME_1:              50,
		DEBUG:                    false,
	}
}

const (
	minProcesses int = 4
	optProcesses int = 6
)

// Don't allow the number of processes to exceed the number of processor
// thread, unless that is smaller than `minProcesses`. If the parameter `n`
// is zero try to return an "optimal" number.
func MaxProcesses(n int) int {
	nmin, np, nopt := MinNpOptProcesses()
	if n == 0 {
		return min(max(nmin, np), nopt)
	}
	if n <= minProcesses {
		return minProcesses
	}
	if n > np {
		return np
	}
	return n
}

func MinNpOptProcesses() (int, int, int) {
	return minProcesses, runtime.NumCPU(), optProcesses
}

/*
One aim is to achieve a – possibly imperfect – timetable within a specified
time. As it is impossible to guarantee that all constraints will be satisfied
within a given time, it may be necessary to drop some of them in order to
place all the activities within this time. However, if it is possible
to satisfy all the constraints within this time, that should be done.

If constraints need to be dropped, there should be an indication of which
ones are difficult. Perhaps more time would help, or a modification of some of
the constraints. Among the difficult ones there may also be constraints which
block the completion of the task and thus _must_ be changed or dropped.

A certain degree of parallel processing is assumed – too few (less than four?)
processor cores is likely to result in a very significant slowdown.

The main function (`StartGeneration`) starts a run with the fully constrained
data and a second run with all the "soft" (non-compulsory) constraints removed.
A further instance is run in which basically all constraints are removed. If
this latter fails, then there is a serious problem with the activities, which
don't fit into the school week. As this unconstrained instance should complete
quickly in most real-life situations, it is given a short timeout. In theory
it is possible that the generation of a timetable could take a longer time
even with the unconstraind data, but such a case would need to be handled in a
different way (and is, in general, a difficult problem ...). This program
assumes that the unconstrained data will allow the rapid placement of all the
activities.

A `TtInstance` structure is constructed to manage the data for each
timetable generation run, each run having its own goroutine. Each instance
has its own individual timeout. There is also a global timeout to stop
all instances which are still running.

Once these initial instances have been started, a "tick-loop" (which is
triggered every second) is entered. This monitors the progress of each active
instance and handles the actions resulting from their completion, whether
successful or not.

Should the fully constrained instance complete successfully within the
allotted time, its result will be saved and all other instances terminated.

When the unconstrained instance completes successfully, a series of further
instances is queued for running, each specifying the addition of a list of
(hard) constraints of a single type. Thus for each type of constraint an
instance is constructed. Using timeouts leading to binary divisions of these
lists an attempt is made to find individual "difficult" constraints, which can
then be disabled in order to get full activity placement within a reasonable
time. Parallel processing can be of some assistance here.

As the constraint types are added one after the other, and often each step
will take longer than the previous one (as the number of constraints grows),
it should be clear why it is desirable that at least the early stages are
completed quickly.

When a single-constraint-type instance completes successfully, it is used as
a new base (`current_instance`) for the addition of further constraints. All
the remaining constraint-type instances are stopped and restarted with this
new base. If a constraint-type instance is timed out, it is stopped and split
into two halves, which then run in its place. If there are no halves (only
one constraint being added) there is no successor and the instance will run
without timeout; should it fail (because of an error or some other halting
criterion, such as "too slow"), the constraint is dropped.

Also when an instance completes successfully within the allotted time, its
result is saved as a `Result` structure, the best result so far gradually
encompassing more of the constraints.

When all the hard constraints have been included or rejected, the soft
constraints are added in basically the same way. If the initial instance with
all hard constraints and no soft constraints completes, this instance will be
used as the new base for adding the soft constraints and the accumulation
loop will be cancelled. If the accumulation loop should finish first (somewhat
unlikely, but possible), the initial instance with all hard constraints may be
terminated.

When everything has been added that can be in the given time, the latest
`Result` is saved as a JSON file, "Result.json". This includes details of the
last successful run and diagnostic information – at least an indication of
which constraints were dropped and any error messages for them which may have
been produced by the generator back-end).
*/

func (attdata *AutoTtData) StartGeneration() {
	base.SetStopFlag(false)

	attdata.lastResult = nil
	attdata.ConstraintErrors = map[ConstraintIndex]string{}
	attdata.instanceCounter = 0
	attdata.current_instance = nil

	// Catch termination signal
	//sigChan := make(chan os.Signal, 1)
	//signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	attdata.active_instances = nil
	attdata.set_runqueue(nil)

	// Global data
	attdata.Ticks = 0

	// First run: all constraints enabled.
	// On successful completion, all other instances should be stopped.
	// If it fails, just this instance should be wound up. Otherwise it
	// should run until the overall time-out, at which point any other active
	// instances should be stopped and the "best" solution at this point
	// chosen.
	{ // Prepare and start fully constrained instance.
		enabled := make([]bool, attdata.NConstraints)
		attdata.log_nconstraints(enabled)
		for i := range attdata.NConstraints {
			// Enable all constraints
			enabled[i] = true
		}
		instance := &TtInstance{
			Index:          -1,
			ConstraintType: "_COMPLETE",
			//Timeout:           0,
			ConstraintEnabled: enabled,
		}
		attdata.full_instance = instance
		attdata.start_instance(instance)
	}

	{ // Prepare instance without soft constraints, enable only the hard constraints.
		// If there are no soft constraints, this is the same as the fully
		// constrained instance
		// This is needed even if SKIP_HARD is set, because it is used as
		// initial base instance for PHASE_SOFT. However, with SKIP_HARD set it
		// will not be run.
		enabled := make([]bool, attdata.NConstraints)
		for _, ilist := range attdata.HardConstraintMap {
			// Enable all hard constraints
			for _, i := range ilist {
				enabled[i] = true
			}
		}
		attdata.hard_instance = &TtInstance{
			Index:          -2,
			ConstraintType: "_HARD_ONLY",
			//Timeout:           0,
			ConstraintEnabled: enabled,
		}
	}

	{ // Prepare instance with only prioritized (hard) constraints.
		// If there aren't any, skip this instance.
		notAvailable := 0
		enabled := make([]bool, attdata.NConstraints)
		for _, natype := range attdata.Source.GetPhase0ConstraintTypes() {
			for _, i := range attdata.HardConstraintMap[natype] {
				notAvailable++
				enabled[i] = true
			}
		}
		if notAvailable != 0 {
			attdata.priority_instance = &TtInstance{
				Index:          -3,
				ConstraintType: "_PRIORITY",
				//Timeout:           0,
				ConstraintEnabled: enabled,
			}
		} else {
			attdata.priority_instance = nil
		}

	}

	{ // Prepare unconstrained instance.
		enabled := make([]bool, attdata.NConstraints)
		attdata.null_instance = &TtInstance{
			Index:          -4,
			ConstraintType: "_UNCONSTRAINED",
			//Timeout:           0 ... attdata.cycle_timeout?,
			ConstraintEnabled: enabled,
		}
	}

	attdata.cycle_timeout = 0

	if len(attdata.HardConstraintMap) == 0 {
		base.LogWarning("--NO_HARD_CONSTRAINTS")
		// But start the instance anyway, either as extra null instance or
		// as basis for SKIP_HARD.
	}
	attdata.start_instance(attdata.hard_instance)
	if TtParameters.SKIP_HARD {
		// Don't run null_instance or priority_instance.
		if len(attdata.SoftConstraintMap) == 0 {
			base.LogWarning("--SKIP_HARD_NO_SOFT")
		}
		attdata.enter_phase(PHASE_SOFT)
	} else {
		if attdata.priority_instance != nil {
			attdata.start_instance(attdata.priority_instance)
		}
		attdata.start_instance(attdata.null_instance)
		// Start in basic phase.
		attdata.enter_phase(PHASE_BASIC)
	}

	// *** Ticker loop ***
	ticker := time.NewTicker(time.Second)

	// The final tidying up – also when an error occurs
	defer func() {
		// Tidy up
		r := recover()
		if r != nil {
			base.LogBug("[%d] !!! RECOVER !!!\n=== %v\n+++\n%s\n---",
				attdata.Ticks, r, debug.Stack())
			fmt.Printf("[%d] !!! RECOVER !!!\n=== %v\n+++\n%s\n---\n",
				attdata.Ticks, r, debug.Stack())
		}
		for {
			// Wait for active instances to finish, stopping them if necessary.
			count := 0
			for _, instance := range attdata.active_instances {
				if !instance.Finished {
					instance.InstanceBackend.DoTick(attdata, instance)
					count++
					attdata.abort_instance(instance)
				}
			}
			if count == 0 {
				ticker.Stop()
				break
			}
			<-ticker.C
		}
		if !TtParameters.DEBUG {
			// Remove all remaining temporary files
			attdata.Backend.Tidy()
		}

		if attdata.current_instance.RunState != INSTANCE_ACCEPTED {
			// If no successes have been booked, there is no result.
			base.LogError("--NO_RESULT")
		} else {
			//TODO: Where (whether?) to save the Result.json file
			jsonbytes := attdata.GetLastResultJSON()
			if len(jsonbytes) != 0 {
				fpath := filepath.Join(base.DataBase.SourceDir, base.DataBase.Name+"_Result.json")
				err := os.WriteFile(fpath, jsonbytes, 0644)
				if err != nil {
					base.LogError("--WRITE_JSON %s", err)
				}
			}
			//TODO: Where (whether?) to save the FET file ...
			// Perhaps return a JSON object containing anything relevant as a field?
			attdata.current_instance.InstanceBackend.FinalizeResult(attdata)
		}

		// Signal end of process
		base.LogCommandEnd(true)
		base.LogTick(-1)
	}()

tickloop:
	for {
		// Remove completed and aborted instances, start queued instances if
		// there are free processors.
		if attdata.update_queue() == 0 {
			base.LogInfo("Run-queue empty")
		}

		<-ticker.C // wait for "tick"

		if base.GetStopFlag() {
			base.LogInfo("!!! INTERRUPTED !!!")
			break tickloop
		}

		attdata.Ticks++
		base.LogTick(attdata.Ticks)

		// Deal with "tick" updates to the `RunState` of the running instances.
		// First increment the ticks of running instances.
		for _, instance := range attdata.active_instances {
			if !instance.Finished {
				// Among other things, update the state:
				instance.InstanceBackend.DoTick(attdata, instance)
			}
		}

		/*/TODO--
		  ilist := []string{}
		  for _, ii := range attdata.get_runqueue() {
		      ilist = append(ilist, fmt.Sprintf("%d:%d", ii.Index, ii.RunState))
		  }
		  base.LogInfo("§Q %+v\n", strings.Join(ilist, ", "))
		  alist := []string{}
		  for _, ii := range attdata.active_instances {
		      alist = append(alist, fmt.Sprintf("%d:%d", ii.Index, ii.RunState))
		  }
		  base.LogInfo("§A %+v\n", strings.Join(alist, ", "))
		*/

		// Then handle the new states
		if attdata.tick_phase() && attdata.phase == PHASE_FINISHED {
			break tickloop
		}

		if attdata.Ticks == TtParameters.TIMEOUT {
			base.LogInfo("!!! TIMEOUT !!!")
			break
		}

	} // tickloop: end
	result := attdata.current_instance
	if result.RunState != INSTANCE_ACCEPTED {
		return // failed
	}
	base.LogInfo("... finalizing ...")

	hnn := 0
	hnall := 0
	snn := 0
	snall := 0
	type constraintinfo struct {
		c string
		n int
		N int
	}
	infolist := []constraintinfo{}

	// Hard constraints sorted by priority
	for _, c := range attdata.Constraint_Types {
		n := 0
		clist := attdata.HardConstraintMap[c]
		for _, cix := range clist {
			if result.ConstraintEnabled[cix] {
				n++
			}
		}
		if len(clist) != 0 {
			infolist = append(infolist,
				constraintinfo{string(c), n, len(clist)})
			hnn += n
			hnall += len(clist)
		}
	}

	// Soft constraints sorted reverse-alphabetically, which puts the
	// highest weights first
	infolist2 := []constraintinfo{}
	for c, clist := range attdata.SoftConstraintMap {
		n := 0
		for _, cix := range clist {
			if result.ConstraintEnabled[cix] {
				n++
			}
		}
		if len(clist) != 0 {
			infolist2 = append(infolist2,
				constraintinfo{string(c), n, len(clist)})
			snn += n
			snall += len(clist)
		}
		slices.SortFunc(infolist2, func(a, b constraintinfo) int {
			return strings.Compare(b.c, a.c)
		})
	}
	infolist = append(infolist, infolist2...)
	for _, info := range infolist {
		base.LogInfo("$ %s: %d / %d", info.c, info.n, info.N)
	}

	report := fmt.Sprintf(
		"::: ALL CONSTRAINTS: (hard) %d / %d  (soft) %d / %d\n",
		hnn, hnall, snn, snall)
	base.LogInfo("%s", report)
}

func (attdata *AutoTtData) start_instance(instance *TtInstance) {
	if instance.Index >= 0 {
		attdata.instanceCounter++
		instance.Index = attdata.instanceCounter
	}
	// Set up enabled constraints (`instance.ConstraintEnabled`) and "timeout".
	if len(instance.Constraints) != 0 {
		// a "normal" instance
		enabled := slices.Clone(instance.BaseInstance.ConstraintEnabled)
		// Add the new constraints
		for _, c := range instance.Constraints {
			enabled[c] = true
		}
		instance.ConstraintEnabled = enabled
		if len(instance.Constraints) == 1 {
			// When only one constraint is being tested, no "timeout"
			instance.Timeout = 0
		} else {
			instance.Timeout = attdata.cycle_timeout
		}
	}
	// Start running.
	instance.RunState = INSTANCE_RUNNING
	attdata.Backend.RunBackend(attdata, instance)
	attdata.active_instances = append(attdata.active_instances, instance)

	//fmt.Printf("?start %d\n", instance.Index)
	//for _, i := range attdata.active_instances {
	//  fmt.Printf("?active %d rs: %d done: %v p: %d\n", i.Index, i.RunState, i.Done, i.Progress)
	//}

}

func (attdata *AutoTtData) abort_instance(instance *TtInstance) {
	if instance != nil && instance.RunState == INSTANCE_RUNNING {
		instance.InstanceBackend.Abort()
		instance.RunState = ABORT_NEW_CYCLE
	}
}
