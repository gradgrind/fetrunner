#include "progress_delegate.h"

ProgressDelegate::ProgressDelegate(QObject *parent)
    : QStyledItemDelegate(parent)
{}

void ProgressDelegate::paint( //
    QPainter *painter,
    const QStyleOptionViewItem &option,
    const QModelIndex &index) const
{
    auto progress = index.data(UserRoleInt).toInt();
    QStyleOptionProgressBar progbar;
    progbar.rect = option.rect;
    progbar.minimum = 0;
    progbar.maximum = 100;
    progbar.progress = progress;
    progbar.text = QString::number(progress).append('%');
    progbar.textVisible = true;
    QApplication::style()->drawControl( //
        QStyle::CE_ProgressBar,
        &progbar,
        painter);
}
