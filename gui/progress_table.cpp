#include "backend.h"
#include "mainwindow.h"
#include "progress_delegate.h"
#include "ui_mainwindow.h"

void init_progress_table() {} //TODO: set up table columns, etc.

void MainWindow::setup_progress_table()
{
    ui->progress_table->setRowCount(0);
    hard_constraint_map.clear();
    soft_constraint_map.clear();
    int i = 0;
    for (const auto &kv : backend->op("HARD_CONSTRAINTS")) {
        auto total = kv.val.toInt();

        //TODO: add table line

        hard_constraint_map[kv.key] = {
            i++,   // index
            0,     // satisfied constraints
            total, // number of constraints
        };
    }
    for (const auto &kv : backend->op("SOFT_CONSTRAINTS")) {
        auto total = kv.val.toInt();

        //TODO: add table line

        soft_constraint_map[kv.key] = {
            i++,   // index
            0,     // satisfied constraints
            total, // number of constraints
        };
    }
}
