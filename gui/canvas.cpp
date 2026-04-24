#include "canvas.h"
#include <QGraphicsItem>
#include <QGraphicsSceneEvent>
#include "chip.h"

const int minutesPerHour = 60;
const qreal CHIP_MARGIN = 1.5;
const qreal CHIP_SPACER = 10.0;

// Unit conversions
const qreal MM2PT = 2.83464549;
const qreal PT2MM = 0.3527778;

// *******************

void on_hover(QGraphicsRectItem* item, bool enter)
{
    qDebug() << (enter ? "ENTER" : "EXIT");
}


// Canvas: This is the "view" widget for the canvas – though it is not itself
// a graphical element! The QGraphicsView handled here is passed in as a
// parameter.
// The actual canvas is implemented as a "scene".
Canvas::Canvas(QGraphicsView *gview) : QObject()
{
    view = gview;
// Change update mode: The default, MinimalViewportUpdate, seems
// to cause artefacts to be left, i.e. it updates too little.
// Also BoundingRectViewportUpdate seems not to be 100% effective.
// view->setViewportUpdateMode(
//     QGraphicsView::ViewportUpdateMode::BoundingRectViewportUpdate
// )
    view->setViewportUpdateMode(
        QGraphicsView::ViewportUpdateMode::FullViewportUpdate
    );
// view->setRenderHints(
//     QPainter::RenderHint::Antialiasing
//     | QPainter::RenderHint::SmoothPixmapTransform
// )
// view->setRenderHints(QPainter::RenderHint::TextAntialiasing)
    view->setRenderHints(QPainter::RenderHint::Antialiasing);

    ldpi = view->logicalDpiX();
    pdpi = view->physicalDpiX();
    // Scaling the scene by pdpi/ldpi should display the correct size ...

    scene = new Scene();
    view->setScene(scene);

    //-- Testing code:
    QGraphicsRectItem *r1 = new QGraphicsRectItem(20, 50, 300, 10);
    scene->addItem(r1);
    scene->addRect(QRectF(200, 300, 100, 100), QPen(Qt::black), QBrush(Qt::red));
    Chip *chip1 = new Chip(200, 40);
    scene->addItem(chip1);
    chip1->setPos(-50, 300);
    chip1->setHoverHandler(on_hover);
    chip1->set_background("f0f000");
    chip1->set_border(3, "ff0000");
    chip1->config_text(14, true, 0, "00E0FF");
    chip1->set_subtext_size(10);
    chip1->set_text("Central");
    chip1->set_toptext("TL", "An extremely long entry which will need shrinking");
    chip1->set_bottomtext("BL", "BR");
    Chip *chip2 = new Chip(40,100);
    scene->addItem(chip2);
    chip2->setPos(-20, 250);
    chip2->set_background("e0e0ff");
    chip2->set_border(0);
    chip2->set_text("Middle");
}

int Canvas::pt2px(int pt) {
    return int(ldpi * pt / 72.0 + 0.5);
}

qreal Canvas::px2mm(int px) {
    return px * 25.4 / ldpi;
}

/*
 * void Canvas::context_1() {
    qDebug() << "-> context_1";
}
*/

// *******************

Scene::Scene() : QGraphicsScene() {}

void Scene::mousePressEvent(QGraphicsSceneMouseEvent *event) {
    if (event->button() == Qt::MouseButton::LeftButton) {
        int kbdmods = qApp->keyboardModifiers();
        int keymod = 0;
        // Note that Ctrl-click is for context menu on OSX ...
        // Note that Alt-click is intercepted by some window managers on
        // Linux ... In that case Ctrl-Alt-click might work.
        if (kbdmods & Qt::KeyboardModifier::AltModifier) {
            keymod = 4;
        } else {
            if (kbdmods & Qt::KeyboardModifier::ShiftModifier) {
                keymod = 1;
            }
            if (kbdmods & Qt::KeyboardModifier::ControlModifier) {
                keymod += 2;
            }
        }
        QPointF point = event->scenePos();

        QList<QGraphicsItem *> allitems = items(point);

        //qDebug() << "Items" << keymod
        //         << " @ " << point << " : " << allitems;

        if (click_handler) {
            click_handler(allitems, keymod);
        }
    }
}

void Scene::set_click_handler(
    std::function<void (const QList<QGraphicsItem *>, int)> handler)
{
    click_handler = handler;
}

void Scene::contextMenuEvent(QGraphicsSceneContextMenuEvent *event)
{
    QPointF point = event->scenePos();
    QList<QGraphicsItem *> allitems = items(point);

    //qDebug() << "Context menu items"
    //         << " @ " << point << " : " << allitems;

    if (context_menu_handler) {
        context_menu_handler(allitems);
    }

}

void Scene::set_context_menu_handler(
    std::function<void (const QList<QGraphicsItem *>)> handler)
{
    context_menu_handler = handler;

    // To show a context menu, cm:
    //    cm->exec(event->screenPos());
    // To make a context menu:
    //    QMenu *cm = new QMenu();
    //    QAction *action = cm->addAction("I am context Action 1");
    //    //connect(action, &QAction::triggered, &Canvas::cm_1);
    //    connect(action, &QAction::triggered, [=](bool b) {
    //        qDebug() << "-> lambda";});
}
