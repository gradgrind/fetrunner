#include "mainwindow.h"
#include <QDir>
#include <QFileDialog>
#include <QMessageBox>
#include "backend.h"
#include "globals.h"
#include "ui_mainwindow.h"

QSettings *settings;
QString file_dir;
QString file_name;
QString file_datatype;
Notifier *notifier;

MainWindow::MainWindow(QWidget *parent)
    : QWidget(parent)
    , ui(new Ui::MainWindow)
{
    ui->setupUi(this);
    notifier = new Notifier();

    ui->help_view->setSource(QUrl("qrc:/help/using_fetrunner.md"));

    backend->registerResultHandler("FETRUNNER_VERSION",
        [this](QString arg) {do_FETRUNNER_VERSION(arg);});
    backend->registerResultHandler("SET_FILE",
        [this](QString arg) {do_SET_FILE(arg);});
    backend->registerResultHandler("DATA_TYPE",
        [this](QString arg) {do_DATA_TYPE(arg);});

    log_view = ui->base_log_view;
    connect( //
        notifier,
        &Notifier::clear_log,
        this,
        &MainWindow::clearLog);
    connect( //
        notifier,
        &Notifier::dump_log,
        this,
        &MainWindow::dumpLog);

    connect( //
        backend,
        &Backend::logcolour,
        this,
        &MainWindow::setLogColour);
    connect( //
        backend,
        &Backend::log,
        this,
        &MainWindow::logLine);

    ttsolver = new FetRunner();
    ui->main_panel->addWidget(ttsolver);

    ttview = new TtView();
    ui->main_panel->addWidget(ttview);

    ttviewselector = new TtViewSelector(ttview);
    ui->side_panel_sub->addWidget(ttviewselector);

    connect( //
        ui->open_file,
        &QPushButton::clicked,
        this,
        &MainWindow::open_file);
    connect( //
        ui->help,
        &QRadioButton::toggled,
        this,
        [this](bool checked) {
            if (checked) {
                ui->main_panel->setCurrentIndex(0);
                ui->side_panel_sub->setCurrentIndex(0);
            }
        });
    connect( //
        ui->general_log,
        &QRadioButton::toggled,
        this,
        [this](bool checked) {
            if (checked) {
                ui->main_panel->setCurrentIndex(1);
                ui->side_panel_sub->setCurrentIndex(0);
            }
        });
    connect( //
        ui->solve_timetable,
        &QRadioButton::toggled,
        this,
        [this](bool checked) {
            if (checked) {
                ui->main_panel->setCurrentIndex(2);
                ui->side_panel_sub->setCurrentIndex(0);
            }
        });
    connect( //
        ui->view_timetable,
        &QRadioButton::toggled,
        this,
        [this](bool checked) {
            if (checked) {
                ui->main_panel->setCurrentIndex(3);
                ui->side_panel_sub->setCurrentIndex(1);
                ttview->enter_view();
            }
        });
    connect( //
        notifier,
        &Notifier::switch_logger,
        this,
        &MainWindow::switch_logger);
    connect( //
        notifier,
        &Notifier::show_logger,
        this,
        &MainWindow::showLogger);
    connect( //
        notifier,
        &Notifier::setBusy,
        this,
        &MainWindow::set_busy);
    connect( //
        notifier,
        &Notifier::errorPopup,
        this,
        &MainWindow::error_popup);
    connect( //
        notifier,
        &Notifier::quit_register_wait,
        this,
        &MainWindow::quit_register_wait);
    connect( //
        notifier,
        &Notifier::finished,
        this,
        &MainWindow::handle_finished);
    connect( //
        notifier,
        &Notifier::new_tt_data,
        this,
        &MainWindow::new_tt_data);
    connect( //
        notifier,
        &Notifier::new_tt_data,
        ttview,
        &TtView::new_tt_data);
    connect( //
        notifier,
        &Notifier::no_tt_data,
        this,
        &MainWindow::no_tt_data);
    connect( //
        notifier,
        &Notifier::fileChanged,
        this,
        &MainWindow::new_file);

    connect( //
        backend,
        &Backend::error,
        this,
        &MainWindow::error_popup);

    settings = new QSettings("gradgrind", "fetrunner");
    const auto geometry = settings->value("gui/MainWindowSize").value<QSize>();
    if (!geometry.isEmpty())
        resize(geometry);
}

MainWindow::~MainWindow()
{
    delete ui;
    settings->setValue("gui/MainWindowSize", size());
    delete settings;
}

void MainWindow::closeEvent(QCloseEvent *e)
{
    //qDebug() << "closeEvent()" << quit_confirmed << quit_requested << waiting_on.length();
    if (quit_confirmed) {
        QWidget::closeEvent(e);
        return;
    }
    if (!quit_requested) {
        quit_requested = true;
        notifier->emit closeRequest();
        // Assume the signal used Qt::DirectConnection, so that all
        // entries in `waiting_on` are now set.
        if (waiting_on.length() == 0) {
            QWidget::closeEvent(e);
            return;
        }
    }
    e->ignore();
}

QTextEdit *MainWindow::selectLogger(int logger) {
    switch (logger) {
    case 0:
        return ui->base_log_view;
    case 1:
       return  ui->solver_log_view;
    case 2:
       return  ui->base_log_view;
    }
    ui->base_log_view->append(QString{"*BUG* Invalid logger index: %1"}.arg(logger));
    emit notifier->show_logger(0);
    return nullptr;
}

void MainWindow::clearLog(int logger) {
    auto lg = selectLogger(logger);
    if (lg != nullptr)
        lg->clear();
}

void MainWindow::dumpLog(int logger) {
    auto lg = selectLogger(logger);
    if (lg == nullptr)
        lg = selectLogger(0);
    QString fname{file_name + ".logdump"};
    QDir fdir{file_dir};
    auto log = lg->toPlainText();
    QFile file(fdir.filePath(fname));
    // Open the file in WriteOnly mode; Truncate to overwrite existing content; Text for line endings
    if (!file.open(QIODevice::WriteOnly | QIODevice::Truncate | QIODevice::Text)) {
        QMessageBox::critical(this, "", file.errorString());
        return;
    }
    // Use QTextStream to write content to the file
    QTextStream out(&file);
    out << log; // Write the input content
    // Optional: Explicitly flush the stream (ensures data is written immediately)
    out.flush();
    // File is automatically closed when 'file' goes out of scope (RAII), but closing explicitly is safe
    file.close();
}

void MainWindow::logLine(QString line) {
    log_view->append(line);
}

void MainWindow::setLogColour(QColor colour) {
    log_view->setTextColor(colour);
}

void MainWindow::quit_register_wait(QString module)
{
    //qDebug() << "quit_register_wait()" << module;
    if (!waiting_on.contains(module)) {
        waiting_on.append(module);
    }
}

void MainWindow::handle_finished(QString module)
{
    //qDebug() << "handle_finished()" << quit_requested;
    if (quit_requested) {
        waiting_on.removeOne(module);
        if (waiting_on.length() == 0) {
            quit_confirmed = true;
            close();
        }
    } else {
        switch_logger("", 0); // revert to base log view
    }
}

void MainWindow::error_popup(const QString msg)
{
    QMessageBox::critical(this, "", msg);
}

void MainWindow::open_file()
{
    //qDebug() << "Open File";

    QString fdir = file_dir;
    if (fdir.isEmpty()) {
        fdir = settings->value("gui/SourceDir", QDir::homePath()).toString();
    }
    QString filepath = QFileDialog::getOpenFileName( //
        this,
        tr("Open Timetable Specification File"),
        fdir,
        tr("FET / W365 Files (*.fet *_w365.json)"));

    if (!filepath.isEmpty()) {
        if (backend->op("SET_FILE", {filepath}))
            notifier->emit fileChanged();
    }
}

void MainWindow::new_file() {
    // Select fetrunner view, disable timetable view.
    no_tt_data();
    ui->solve_timetable->setEnabled(true);
    ui->solve_timetable->click();
}

void MainWindow::set_busy(bool on) {
    ui->open_file->setDisabled(on);
    //ui->control_panel->setDisabled(on);
}

void MainWindow::showLogger(int logger) {
    ui->log_tabs->setCurrentIndex(logger);
    ui->general_log->click();
}

void MainWindow::switch_logger(QString msg, int log_viewer) {
    if (!msg.isEmpty())
        ui->base_log_view->append(msg);
    switch (log_viewer) {
    case 0:
        log_view = ui->base_log_view;
        break;
    case 1:
        log_view = ui->solver_log_view;
        break;
    case 2:
        log_view = ui->timetable_log_view;
        break;
    default:
        log_view = ui->base_log_view;
        log_view->append("*BUG* Invalid log viewer: " + QString::number(log_viewer));
        log_viewer = 0;
    }
    ui->log_tabs->setCurrentIndex(log_viewer);
}

void MainWindow::do_FETRUNNER_VERSION(const QString &val) {
    ui->fetrunner_version->setText(val);
}

void MainWindow::do_SET_FILE(const QString &val) {
    QDir dir{val};
    file_name = dir.dirName();
    dir.cdUp();
    ui->file_path->setText(dir.absoluteFilePath(file_name));
    auto fdir = dir.absolutePath();
    if (fdir != file_dir) {
        file_dir = fdir;
        settings->setValue("gui/SourceDir", fdir);
    }
}

void MainWindow::do_DATA_TYPE(const QString &val) {
    file_datatype = val;
}

void MainWindow::new_tt_data() {
    if (quit_requested) return;
    ui->view_timetable->setEnabled(true);
}

void MainWindow::no_tt_data() {
    ui->view_timetable->setEnabled(false);
}
