#ifndef PROGRESS_DELEGATE_H
#define PROGRESS_DELEGATE_H

#include <QApplication>
#include <QStyledItemDelegate>

class ProgressDelegate : public QStyledItemDelegate
{
public:
    ProgressDelegate(QObject *parent = nullptr);

    // Override this function to paint the progress bar.
    void paint(QPainter *painter,
               const QStyleOptionViewItem &option,
               const QModelIndex &index) const override;
};

#endif // PROGRESS_DELEGATE_H
