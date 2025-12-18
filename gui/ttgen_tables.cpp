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
    hard_constraint_map.clear();
    soft_constraint_map.clear();
    auto row = ui->progress_table->rowCount();
    for (const auto &kv : backend->op("HARD_CONSTRAINTS")) {
        auto cname = constraint_name(kv.key);
        // add table line
        auto item0 = new QTableWidgetItem("[!] " + cname);  // constraint type
        auto item1 = new QTableWidgetItem("/ " + kv.val);   // number of constraints
        auto item2 = new QTableWidgetItem("0");             // accepted constraints
        auto item3 = new QTableWidgetItem("@ 0");           // number of constraints
        ui->progress_table->insertRow(row);
        ui->progress_table->setItem(row, 0, item2);
        ui->progress_table->setItem(row, 1, item1);
        ui->progress_table->setItem(row, 2, item3);
        ui->progress_table->setItem(row, 3, item0);

        hard_constraint_map[cname] = {
            row++,          // index
            0,              // satisfied constraints
            kv.val.toInt(), // number of constraints
        };
    }
    if (hard_constraint_map.size() != 0) {
        ui->label_hard->setEnabled(true);
        ui->progress_hard->setEnabled(true);
    }
    for (const auto &kv : backend->op("SOFT_CONSTRAINTS")) {
        auto cname = constraint_name(kv.key);
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

        soft_constraint_map[cname] = {
            row++,          // index
            0,              // satisfied constraints
            kv.val.toInt(), // number of constraints
        };
    }
    if (soft_constraint_map.size() != 0) {
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

void MainWindow::tableProgress(instance_row &irow)
{
    auto constraint = irow.data[1];
    auto number = irow.data[2];
    if (!irow.data[4].isEmpty()) { // hard constraint
        if (!hard_constraint_map.contains(constraint))
            qFatal() << "hard_constraint_map, no key" << constraint;
        progress_line &cdata = hard_constraint_map[constraint];
        cdata.progress += number.toInt();
        if (cdata.progress == cdata.total)
            ui->progress_table->item(cdata.index, 0)->setText("+++");
        else if (cdata.progress > cdata.total)
            qFatal() << "cdata.progress > cdata.total" << "(hard)" << constraint;
        else
            ui->progress_table->item(cdata.index, 0)->setText(QString::number(cdata.progress));
        ui->progress_table->item(cdata.index, 2)->setText("@ " + timeTicks);
    } else {
        if (!soft_constraint_map.contains(constraint))
            qFatal() << "soft_constraint_map, no key" << constraint;
        progress_line &cdata = soft_constraint_map[constraint];
        cdata.progress += number.toInt();
        if (cdata.progress == cdata.total)
            ui->progress_table->item(cdata.index, 0)->setText("+++");
        if (cdata.progress > cdata.total)
            qFatal() << "cdata.progress > cdata.total" << "(soft)" << constraint;
        else
            ui->progress_table->item(cdata.index, 0)->setText(QString::number(cdata.progress));
        ui->progress_table->item(cdata.index, 2)->setText("@ " + timeTicks);
    }
}

void MainWindow::tableProgressAll()
{
    tableProgressHard();
    tableProgressGroup(soft_constraint_map);
}

void MainWindow::tableProgressHard()
{
    tableProgressGroup(hard_constraint_map);
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

void MainWindow::instanceRowProgress(int key, QStringList parms)
{
    // If the entry is not in the map, add a new entry.
    auto irow = instance_row_map.value(key);
    int row;
    if (irow.item == nullptr) {
        auto ctype = irow.data[1]; // constraint type
        if (!irow.data[4].isEmpty())
            ctype.prepend("[!] "); // hard constraint
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

        QTimer::singleShot(0, [this, item4]() { //
            this->ui->instance_table->scrollToItem(item4);
        });

    } else {
        row = ui->instance_table->row(irow.item);
    }

    irow.item->setText(parms[1] + "%");                         // progress (%)
    ui->instance_table->item(row, 2)->setText("@ " + parms[2]); // @ time
}
