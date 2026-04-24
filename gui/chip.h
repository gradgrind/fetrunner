#ifndef CHIP_H
#define CHIP_H

#include <QGraphicsRectItem>
#include <QGraphicsScene>
#include <QGraphicsSceneMouseEvent>
#include <QRegularExpression>
//#include <QJsonObject>

// *******************

class HoverRectItem : public QGraphicsRectItem
{
public:
    enum { Type = UserType + 1 };
    int type() const override
    {
        // Enable the use of qgraphicsitem_cast with this item.
        return Type;
    }

    HoverRectItem(QGraphicsItem *parent = nullptr);
    void setHoverHandler(
        std::function<void (HoverRectItem *, bool)> handler);

protected:
    void hoverEnterEvent(QGraphicsSceneHoverEvent *event) override;
    void hoverLeaveEvent(QGraphicsSceneHoverEvent *event) override;

private:
    std::function<void (HoverRectItem*, bool)> hover_handler;
};

// *******************

class Chip : public HoverRectItem
{
public:
    enum { Type = UserType + 2 };
    int type() const override
    {
        // Enable the use of qgraphicsitem_cast with this item.
        return Type;
    }

    Chip();
    Chip(qreal width, qreal height);

    virtual void set_size(qreal width, qreal height);
    void set_background(QString colour);
    void set_border(qreal width, QString colour = "000000");
    void config_text(
        qreal tsize,
        bool tbold = false,
        int align = 0,
        QString colour = "");
    void set_subtext_size(qreal tsize);
    void set_text(QString text);
    void set_toptext(QString text_l, QString text_r);
    void set_bottomtext(QString text_l, QString text_r);
    QMenu *context_menu = nullptr;

private:
    void set_item(QGraphicsSimpleTextItem *&item, QString text, QFont font);
    QGraphicsSimpleTextItem *m_item = nullptr;
    QGraphicsSimpleTextItem *tl_item = nullptr;
    QGraphicsSimpleTextItem *tr_item = nullptr;
    QGraphicsSimpleTextItem *bl_item = nullptr;
    QGraphicsSimpleTextItem *br_item = nullptr;
    void place_pair(
        QGraphicsSimpleTextItem *l,
        QGraphicsSimpleTextItem *r,
        bool top);

    QFont central_font;
    int central_align = 0; // <0 => left, 0 => centre, >0 => right
    QString central_colour;
    QFont corner_font;
};

static const QRegularExpression re_colour("^[0-9a-fA-F]{6}$");

#endif // CHIP_H
