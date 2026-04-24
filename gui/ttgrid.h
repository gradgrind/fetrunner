#ifndef TTGRID_H
#define TTGRID_H

#include <QApplication>
#include "canvas.h"
#include "chip.h"
//#include "database.h"
#include <QStringList>
#include <qjsonobject.h>

class Cell : public Chip
{
public:
    enum { Type = UserType + 3 };
    int type() const override
    {
        // Enable the use of qgraphicsitem_cast with this item.
        return Type;
    }

    Cell(int x, int y);

    int cellx;
    int celly;
};

class Tile; // forward declaration

enum HighlightColour { NOCLASH = 0, ONLY_FLEXIROOM, REPLACEABLE };

class TtGrid
{
public:
    TtGrid( //
        QGraphicsView *view,
        QStringList days,
        QStringList hours,
        QList<int> breaks = {});
    ~TtGrid();

    void setup_grid();
    void place_tile(Tile *tile, int col, int row);
    void setClickHandler(std::function<void (int, int, Tile *, int)> handler);
    void select_tile(Tile *tile);
    void clearHighlights();
    void setHighlight(int day, int hour, HighlightColour colour);

    Canvas *canvas;
    Scene *scene;
    QStringList daylist;
    QStringList hourlist;
    QList<int> breaklist;

    QList<QList<Cell *>> cols;

    const qreal DAY_WIDTH = 140.0;
    const qreal HOUR_HEIGHT = 60.0;
    const qreal VHEADERWIDTH = 80.0;
    const qreal HHEADERHEIGHT = 40.0;
    const qreal GRIDLINEWIDTH = 2.0;
    const QString GRIDLINECOLOUR = "EDAB9A";
    const QString BREAKLINECOLOUR = "404080";
    const qreal FONT_CENTRE_SIZE = 12.0;
    const qreal FONT_CORNER_SIZE = 8.0;
    const QString SELECTIONCOLOUR = "FF0000";
    const QBrush OKBRUSH = QColor(255, 223, 255, 255);

    QJsonObject settings;

    std::function<void (HoverRectItem*, bool)> hover_handler;

    QHash<int, Tile *> lid2tile;

private:
    void handle_click(QList<QGraphicsItem *> items, int keymod);
    void handle_context_menu(QList<QGraphicsItem *> items);
    void handle_hover(HoverRectItem*, bool);
    std::function<void (int day, int hour, Tile *tile, int keymode)>
        click_handler;
    QGraphicsRectItem *selection_rect = nullptr;
    QList<Cell*> ok_cells;
    QList<QList<QGraphicsRectItem *>> highlights;

    const QList<QColor> HighlightColours{QColor(11, 255, 39, 128),
                                         QColor(99, 148, 255, 128),
                                         QColor(255, 80, 29, 128)};
};

class Tile : public Chip
{
public:
    enum { Type = UserType + 4 };
    int type() const override
    {
        // Enable the use of qgraphicsitem_cast with this item.
        return Type;
    }

    Tile(TtGrid *grid, QJsonObject data, int lesson_id);

    void place(qreal x, qreal y, qreal w, qreal h);

    int lid;
    QString ref;
    int length;
    int divs;
    int div0;
    int ndivs;
    QString middle;
    QString tl;
    QString tr;
    QString bl;
    QString br;

    const qreal TILE_BORDER_WIDTH = 1.0;
    const QString TILE_BORDER_COLOUR = "6060FF";
    const bool TEXT_BOLD = true;
    const int TEXT_ALIGN = 0; // centred
};

#endif // TTGRID_H

