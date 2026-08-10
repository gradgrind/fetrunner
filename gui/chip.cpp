#include "chip.h"
#include <QBrush>
#include <QPen>
#include <QFont>

HoverRectItem::HoverRectItem(QGraphicsItem *parent)
    : QGraphicsRectItem(parent) {}

void HoverRectItem::setHoverHandler(
    std::function<void (HoverRectItem *, bool)> handler)
{
    hover_handler = handler;
    if (handler) {
        setAcceptHoverEvents(true);
    } else {
        setAcceptHoverEvents(false);
    }
}

void HoverRectItem::hoverEnterEvent(QGraphicsSceneHoverEvent *event) {
    hover_handler(this, true);
}

void HoverRectItem::hoverLeaveEvent(QGraphicsSceneHoverEvent *event) {
    hover_handler(this, false);
}

// *******************

/** A rectangular box with border colour, border width and background colour.
 *  The default fill is none (transparent), the default pen is a black
 *  line with width = 1 (the width of a <QPen> can be set to an <int> or
 *  a <float>).
 *  The item's coordinate system starts at (0, 0), fixed by passing
 *  this origin to the <QGraphicsRectItem> constructor.
 *  The box is then moved to the desired location using method "setPos".
 *  It can have a vertically centred simple text item, which can be aligned
 *  horizontally left, centre or right, also a simple text item in each
 *  of the four corners:
 *      "tl" – top left     "tr" – top right
 *      "bl" – bottom left  "br" – bottom right
 *  The font and colour of the centred text can be set separately from
    those of the corners.
*/
Chip::Chip(qreal width, qreal height) : HoverRectItem() {
    set_size(width, height);
}

Chip::Chip() : HoverRectItem() {}

void Chip::set_size(qreal width, qreal height)
{
    setRect(0.0, 0.0, width, height);
}

// setPos(x, y); built in member function

/* Colour the background, which is initially transparent.
 * Colours must be provided as "RRGGBB" strings (case insensitive).
 * The chip becomes opaque.
*/
void Chip::set_background(QString colour)
{
    if (re_colour.match(colour).hasMatch()) {
        setBrush(QBrush(QColor("#FF" + colour)));
    } else {
        qFatal("Invalid background colour: %s", qUtf8Printable(colour));
    }
}

/* Set the border width and colour, which is initially black with width = 1.
 * Colours must be provided as "RRGGBB" strings (case insensitive).
*/
void Chip::set_border(qreal width, QString colour) {
    if (width > 0.01) {
        if (re_colour.match(colour).hasMatch()) {
            setPen(QPen(QBrush(QColor("#FF" + colour)), width));
        } else {
            qFatal("Invalid border colour: %s", qUtf8Printable(colour));
        }
    } else {
        setPen(Qt::PenStyle::NoPen);
    }
}

const qreal CHIP_MARGIN = 1.5;

/* Set the text items within the chip.
 * This also supports rewriting the chip's text.
 * Unfortunately C++ doesn't support passing parameters by name,
 * so there are various methods to set up these text items.
*/

void Chip::config_text(qreal tsize, bool tbold, int align, QString colour)
{
    QFont f;
    if (tsize > 0.01) {
        f.setPointSizeF(tsize);
    }
    f.setBold(tbold);
    central_font = f;
    central_align = align;
    if (re_colour.match(colour).hasMatch()) {
        central_colour = "#FF" + colour;
    }
}

void Chip::set_subtext_size(qreal tsize) {
    QFont f;
    if (tsize > 0.01) {
        f.setPointSizeF(tsize);
    }
    corner_font = f;
}

void Chip::set_text(QString text) {
    set_item(m_item, text, central_font);
    if (m_item) {
        if (!central_colour.isEmpty()) {
            m_item->setBrush(QBrush(QColor(central_colour)));
        }
        // Get chip dimensions
        QRectF bb = rect();
        qreal h0 = bb.height();
        qreal w0 = bb.width();
        // Measure the text
        bb = m_item->boundingRect();
        qreal h = bb.height();
        qreal w = bb.width();
        qreal scale = (w0 - CHIP_MARGIN * 2) * 0.9 / w;
        if (scale < 1.0) {
            //TODO: use text_truncate
            w *= scale;
            h *= scale;
        } else {
            scale = 1.0;
        }
        m_item->setScale(scale);
        // Deal with alignment
        qreal x;
        if (central_align < 0) {
            // align left
            x = CHIP_MARGIN;
        } else if (central_align == 0) {
            // centred
            x = (w0 - w) / 2;
        } else {
            // align right
            x = w0 - CHIP_MARGIN - w;
        }
        m_item->setPos(x, (h0 - h) / 2);
    }
}

const qreal SCALE_LIMIT = 0.5;

QString text_truncate(QString text, qreal scale)
{
    if (scale < SCALE_LIMIT) {
        qsizetype i0 = text.length() * 2 * scale;
        auto i = text.lastIndexOf(',', i0);
        if (i > 0) return text.first(i) + ",+";
        return text.first(i0) + "+";
    }
    return {};
}


void Chip::set_item(QGraphicsSimpleTextItem *&item, QString text, QFont font)
{
    if (text.isEmpty()) {
        if (item) {
            scene()->removeItem(item);
            delete item;
            item = nullptr;
        }
        return;
    }
    if (!item) {
        item = new QGraphicsSimpleTextItem(this);
    }
    item->setText(text);
    item->setFont(font);
}

void Chip::set_toptext(QString text_l, QString text_r) {
    set_item(tl_item, text_l, corner_font);
    set_item(tr_item, text_r, corner_font);
    place_pair(tl_item, tr_item, true);
}

void Chip::set_bottomtext(QString text_l, QString text_r) {
    set_item(bl_item, text_l, corner_font);
    set_item(br_item, text_r, corner_font);
    place_pair(bl_item, br_item, false);
}

void Chip::place_pair(
    QGraphicsSimpleTextItem *&l,
    QGraphicsSimpleTextItem *&r,
    bool top)
{
    qreal w0 = rect().width() - CHIP_MARGIN * 2;
    auto w0_lr = w0 * 0.8;
    qreal w_l = 0.0;
    int lr = 0;
    if (l) {
        // Reset scale
        w_l = l->boundingRect().width();
        lr += 1;
    }
    qreal w_r = 0.0;
    if (r) {
        // Reset scale
        w_r = r->boundingRect().width();
        lr += 2;
    }

    //TODO: Handle null items
    if (lr == 0) return;

    // Get scales
    qreal s_l = 1.0;
    qreal s_r = 1.0;

    qreal s0 = w0_lr / (w_l + w_r);
    if (s0 < SCALE_LIMIT) {
        // Need some truncating
        auto f_l = w_l / w_r;
        auto f_r = 1 - f_l;
        if (lr & 1) {
            auto s0_l = f_l * w0_lr / w_l;
            if (s0_l < 1) {
                auto t1 = text_truncate(l->text(), s0_l);
                if (t1.isEmpty()) {
                    s_l = s0_l;
                } else {
                    qDebug() << "§l" << t1;
                    set_item(l, t1, corner_font);
                    w_l = l->boundingRect().width();
                    s_l = f_l * w0_lr / w_l;
                }
            }
        }
        if (lr & 2) {
            auto s0_r = f_r * w0_lr / w_r;
            if (s0_r < 1) {
                auto t1 = text_truncate(r->text(), s0_r);
                if (t1.isEmpty()) {
                    s_r = s0_r;
                } else {
                    qDebug() << "§l" << t1;
                    set_item(r, t1, corner_font);
                    w_r = r->boundingRect().width();
                    s_r = f_r * w0_lr / w_r;
                }
            }
        }
    } else if (s0 < 1.0) {
        // Need scaling without truncation
        s_l = s0;
        s_r = s0;
    }
    // Place items
    if (l) {
        l->setScale(s_l);
        if (top) {
            l->setPos(CHIP_MARGIN, CHIP_MARGIN);
        } else {
            qreal h0 = rect().height() - CHIP_MARGIN;
            l->setPos(CHIP_MARGIN,
                      h0 - l->sceneBoundingRect().height());
        }
    }
    if (r) {
        r->setScale(s_r);
        w_r *= s_r;
        if (top) {
            r->setPos(w0 + CHIP_MARGIN - w_r, CHIP_MARGIN);
        } else {
            qreal h0 = rect().height() - CHIP_MARGIN;
            r->setPos(w0 + CHIP_MARGIN - w_r,
                      h0 - r->sceneBoundingRect().height());
        }
    }
}
