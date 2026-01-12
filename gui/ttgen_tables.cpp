#include <QDir>
#include <QTimer>
#include "backend.h"
#include "mainwindow.h"
#include "ui_mainwindow.h"

void MainWindow::init_ttgen_tables()
{
    //ui->instance_table->horizontalHeader()->setSectionResizeMode(QHeaderView::Fixed);
    //ui->instance_table->resizeColumnsToContents();
    QFontMetrics fm(ui->instance_table->font());
    //int w = 0;
    for (auto col = 0; col < 4; ++col) {
        auto headerItem = ui->instance_table->horizontalHeaderItem(col);
        auto text = headerItem->text();
        int col_width = fm.horizontalAdvance(text) + 10; // add some padding
        if (col == 3) {
            auto wmin = fm.horizontalAdvance("00000000000000");
            if (col_width < wmin)
                col_width = wmin;
        }
        //w += col_width;
        ui->instance_table->setColumnWidth(col, col_width);
    }

    /*
    int wsb = qApp->style()->pixelMetric(QStyle::PM_ScrollBarExtent);
    instance_table_fixed_width = w + wsb + 10;
    ui->instance_table->setItemDelegateForColumn( //
        3,
        new ProgressDelegate(ui->instance_table));
    ui->progress_table->setItemDelegateForColumn( //
        2,
        new ProgressDelegate(ui->progress_table));
    */
}

void MainWindow::setup_progress_table()
{
    //TODO--hard_constraint_map.clear();
    //TODO--soft_constraint_map.clear();
    constraint_map.clear();
    auto row = ui->progress_table->rowCount();
    for (const auto &kv : backend->op("HARD_CONSTRAINTS")) {
        //auto cname = constraint_name(kv.key);
        auto cname = kv.key;
        // add table line
        //auto item0 = new QTableWidgetItem("[!] " + cname);  // constraint type
        auto item0 = new QTableWidgetItem(cname);           // constraint type
        auto item1 = new QTableWidgetItem("/ " + kv.val);   // number of constraints
        auto item2 = new QTableWidgetItem("0");             // accepted constraints
        auto item3 = new QTableWidgetItem("@ 0");           // number of constraints
        ui->progress_table->insertRow(row);
        ui->progress_table->setItem(row, 0, item2);
        ui->progress_table->setItem(row, 1, item1);
        ui->progress_table->setItem(row, 2, item3);
        ui->progress_table->setItem(row, 3, item0);

        //TODO--hard_constraint_map[cname] = {
        constraint_map[cname] = {
            //
            row++,          // index
            0,              // satisfied constraints
            kv.val.toInt(), // number of constraints
        };
    }
    //TODO--if (hard_constraint_map.size() != 0) {
    auto hcmapsize = constraint_map.size();
    if (hcmapsize != 0) {
        ui->label_hard->setEnabled(true);
        ui->progress_hard->setEnabled(true);
    }
    for (const auto &kv : backend->op("SOFT_CONSTRAINTS")) {
        //auto cname = constraint_name(kv.key);
        auto cname = kv.key;
        // add table line
        auto item0 = new QTableWidgetItem(cname);         // constraint type
        auto item1 = new QTableWidgetItem("/ " + kv.val); // number of constraints
        auto item2 = new QTableWidgetItem("0");           // accepted constraints
        auto item3 = new QTableWidgetItem("@ 0");         // number of constraints
        ui->progress_table->insertRow(row);
        ui->progress_table->setItem(row, 0, item2);
        ui->progress_table->setItem(row, 1, item1);
        ui->progress_table->setItem(row, 2, item3);
        ui->progress_table->setItem(row, 3, item0);

        //TODO--soft_constraint_map[cname] = {
        constraint_map[cname] = {
            //
            row++,          // index
            0,              // satisfied constraints
            kv.val.toInt(), // number of constraints
        };
    }
    //TODO--if (soft_constraint_map.size() != 0) {
    if (constraint_map.size() != hcmapsize) {
        ui->label_soft->setEnabled(true);
        ui->progress_soft->setEnabled(true);
    }
}

void MainWindow::nconstraints(const QString &data)
{
    auto slist = data.split(u'.');
    auto h = slist[0];
    auto hn = slist[1];
    auto s = slist[2];
    auto sn = slist[3];
    if (h != hard_count) {
        // If `hn` is zero ("0"), this will only be run once.
        ui->hard_naccepted->setText(h);
        ui->hard_tlastchange->setText(timeTicks);
        if (hard_count.isEmpty()) {
            // the first call
            ui->hard_nconstraints->setText(hn);
        }
        hard_count = h;
        auto hi = hn.toInt();
        if (hi != 0)
            ui->progress_hard->setValue((h.toInt() * 100) / hi);
        else
            ui->progress_hard->setValue(-1);
    } else if (h != "0") {
        // All hard constraints fulfilled, ensure that progress
        // table reflects this.
        //TODO--tableProgressHard();
        tableProgressSet(true);
    }
    if (s != soft_count) {
        // If `sn` is zero ("0"), this will only be run once.
        ui->soft_naccepted->setText(s);
        ui->soft_tlastchange->setText(timeTicks);
        if (soft_count.isEmpty()) {
            // the first call
            ui->soft_nconstraints->setText(sn);
        }
        soft_count = s;
        auto si = sn.toInt();
        if (si != 0)
            ui->progress_soft->setValue((s.toInt() * 100) / si);
        else
            ui->progress_soft->setValue(-1);
    }
}

bool MainWindow::dump_log(QString fname)
{
    QDir fdir{ui->currentDir->text()};
    auto log = ui->logview->toPlainText();
    QFile file(fdir.filePath(fname));
    // Open the file in WriteOnly mode; Truncate to overwrite existing content; Text for line endings
    if (!file.open(QIODevice::WriteOnly | QIODevice::Truncate | QIODevice::Text)) {
        qDebug() << "Failed to open file for writing:" << file.errorString();
        return false; // Indicate failure
    }
    // Use QTextStream to write content to the file
    QTextStream out(&file);
    out << log; // Write the input content
    // Optional: Explicitly flush the stream (ensures data is written immediately)
    out.flush();
    // File is automatically closed when 'file' goes out of scope (RAII), but closing explicitly is safe
    file.close();
    return true;
}

void MainWindow::fail(QString msg)
{
    dump_log(ui->currentFile->text() + ".logdump");

    close();
    QMessageBox::critical(this, "", msg);
    qApp->quit();
}

void MainWindow::tableProgress(progress_changed update)
{
    auto constraint = update.constraint;
    auto delta = update.number.toInt();

    if (!constraint_map.contains(constraint)) {
        fail("*BUG* constraint_map, no key " + constraint);
        return;
    }
    progress_line &cdata = constraint_map[constraint];
    cdata.progress += delta;
    if (cdata.progress == cdata.total) {
        ui->progress_table->item(cdata.index, 0)->setText("+++");
    } else if (cdata.progress > cdata.total) {
        ui->logview->append(QString{"\n***DUMP*** %1 %2 %3 %4\n"}
                                .arg(constraint)
                                .arg(cdata.progress)
                                .arg(delta)
                                .arg(cdata.total));

        fail("*BUG* cdata.progress > cdata.total " + constraint);
        return;
    } else {
        ui->progress_table->item(cdata.index, 0)->setText(QString::number(cdata.progress));
    }
    ui->progress_table->item(cdata.index, 2)->setText("@ " + timeTicks);

    /*
    if (!constraint.contains(':')) { // hard constraint
        if (!hard_constraint_map.contains(constraint)) {
            fail("*BUG* hard_constraint_map, no key " + constraint);
            return;
        }
        progress_line &cdata = hard_constraint_map[constraint];
        cdata.progress += delta;
        if (cdata.progress == cdata.total) {
            ui->progress_table->item(cdata.index, 0)->setText("+++");
        } else if (cdata.progress > cdata.total) {
            ui->logview->append(QString{"\n***DUMP*** %1 %2 %3 %4\n"}
                                    .arg(constraint)
                                    .arg(cdata.progress)
                                    .arg(delta)
                                    .arg(cdata.total));

            fail("*BUG* cdata.progress > cdata.total (hard) " + constraint);
            return;
        } else
            ui->progress_table->item(cdata.index, 0)->setText(QString::number(cdata.progress));
        ui->progress_table->item(cdata.index, 2)->setText("@ " + timeTicks);
    } else {
        if (!soft_constraint_map.contains(constraint)) {
            fail("*BUG* soft_constraint_map, no key " + constraint);
            return;
        }
        progress_line &cdata = soft_constraint_map[constraint];
        cdata.progress += delta;
        if (cdata.progress == cdata.total)
            ui->progress_table->item(cdata.index, 0)->setText("+++");
        else if (cdata.progress > cdata.total) {
            ui->logview->append(QString{"\n***DUMP*** %1 %2 %3 %4\n"}
                                    .arg(constraint)
                                    .arg(cdata.progress)
                                    .arg(delta)
                                    .arg(cdata.total));

            fail("*BUG* cdata.progress > cdata.total (soft) " + constraint);
            return;
        } else
            ui->progress_table->item(cdata.index, 0)->setText(QString::number(cdata.progress));
        ui->progress_table->item(cdata.index, 2)->setText("@ " + timeTicks);
    }
*/
}

/*
//TODO: Do I need to change where this is called from (should
// probably be from ticker).
void MainWindow::tableProgressAll()
{
    tableProgressHard();
    tableProgressGroup(soft_constraint_map);
    //TODO: set number and progress bar (soft)??
    // Actually, isn't that done by .NCONSTRAINTS?
    //ui->progress_soft->setValue(100);
}

//TODO: Do I need to change where this is called from (should
// probably be from ticker).
void MainWindow::tableProgressHard()
{
    tableProgressGroup(hard_constraint_map);
    //TODO: set number and progress bar (hard).
    // Actually, isn't that done by .NCONSTRAINTS?
    //ui->progress_hard->setValue(100);
}

void MainWindow::tableProgressGroup(QHash<QString, progress_line> hsmap)
{
    for (auto it = hsmap.begin(); it != hsmap.end(); ++it) {
        progress_line &cdata = it.value();
        if (cdata.progress != cdata.total) {
            ui->progress_table->item(cdata.index, 0)->setText("+++");
            cdata.progress = cdata.total;
            ui->progress_table->item(cdata.index, 2)->setText("@ " + timeTicks);
        }
    }
}
*/

void MainWindow::tableProgressSet(bool hard_only)
{
    for (auto it = constraint_map.begin(); it != constraint_map.end(); ++it) {
        if (hard_only && it.key().contains(':'))
            continue;
        progress_line &cdata = it.value();
        if (cdata.progress != cdata.total) {
            ui->progress_table->item(cdata.index, 0)->setText("+++");
            cdata.progress = cdata.total;
            ui->progress_table->item(cdata.index, 2)->setText("@ " + timeTicks);
        }
    }
}

void MainWindow::instanceRowProgress(int key, QStringList parms)
{
    // If the entry is not in the map, add a new entry.
    auto irow = instance_row_map.value(key);
    int row;
    if (irow.item == nullptr) {
        auto ctype = irow.data[1]; // constraint type
        auto item0 = new QTableWidgetItem(ctype);
        auto item1 = new QTableWidgetItem(irow.data[2]); // number of constraints
        item1->setTextAlignment(Qt::AlignCenter);
        auto timeout = irow.data[3]; // timeout
        if (timeout == "0")
            timeout = "---";
        else
            timeout.prepend("/ ");
        auto item2 = new QTableWidgetItem(timeout);
        item2->setTextAlignment(Qt::AlignCenter);
        auto item3 = new QTableWidgetItem(); // @ time
        item3->setTextAlignment(Qt::AlignCenter);
        auto item4 = new QTableWidgetItem(); // progress (%)
        row = ui->instance_table->rowCount();
        ui->instance_table->insertRow(row);
        ui->instance_table->setItem(row, 0, item4);
        ui->instance_table->setItem(row, 1, item1);
        ui->instance_table->setItem(row, 2, item3);
        ui->instance_table->setItem(row, 3, item2);
        ui->instance_table->setItem(row, 4, item0);
        irow.item = item4;
        instance_row_map[key] = irow;

        // The table widget is scrolled to the bottom on each tick
        // (see MainWindow::ticker()).

    } else {
        row = ui->instance_table->row(irow.item);
    }

    irow.item->setText(parms[1] + "%");                         // progress (%)
    ui->instance_table->item(row, 2)->setText("@ " + parms[2]); // @ time
}
