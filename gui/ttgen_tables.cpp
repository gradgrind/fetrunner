#include <QTimer>
#include "backend.h"
#include "mainwindow.h"
#include "progress_delegate.h"
#include "ui_mainwindow.h"

void MainWindow::init_ttgen_tables()
{
    //ui->instance_table->horizontalHeader()->setSectionResizeMode(QHeaderView::Fixed);
    //ui->instance_table->resizeColumnsToContents();
    QFontMetrics fm(ui->instance_table->font());
    int w = 0;
    for (auto col = 0; col < 4; ++col) {
        auto headerItem = ui->instance_table->horizontalHeaderItem(col);
        auto text = headerItem->text();
        int col_width = fm.horizontalAdvance(text) + 10; // add some padding
        if (col == 3) {
            auto wmin = fm.horizontalAdvance("00000000000000");
            if (col_width < wmin)
                col_width = wmin;
        }
        w += col_width;
        ui->instance_table->setColumnWidth(col, col_width);
    }
    int wsb = qApp->style()->pixelMetric(QStyle::PM_ScrollBarExtent);
    instance_table_fixed_width = w + wsb + 10;
    ui->instance_table->setItemDelegateForColumn( //
        3,
        new ProgressDelegate(ui->instance_table));
    ui->progress_table->setItemDelegateForColumn( //
        2,
        new ProgressDelegate(ui->progress_table));
}

void MainWindow::setup_progress_table()
{
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

void MainWindow::tableProgress() {}
