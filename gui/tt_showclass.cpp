#include "tt_showclass.h"
#include <QJsonArray>

ShowClass::ShowClass(TtGrid *grid, TimetableData *tt_data, int class_id)
{
    DBData *db_data = tt_data->db_data;
    for (int course_id : tt_data->class_courses[class_id]) {
        const auto course = db_data->Nodes.value(course_id);
        QStringList teachers;
        const auto tlist = course.value("TEACHERS").toArray();
        for (const auto & t : tlist) {
            teachers.append(db_data->get_tag(t.toInt()));
        }
        QString teacher = teachers.join(",");
        QString subject = db_data->get_tag(course.value("SUBJECT").toInt());
        QStringList rooms;
        const auto rlist = course.value("FIXED_ROOMS").toArray();
        for (const auto & r : rlist) {
            rooms.append(db_data->get_tag(r.toInt()));
        }
        const auto tile_info = tt_data->course_tileinfo[course_id];
        const auto tiles = tile_info.value(class_id);
        for (int lid : db_data->course_lessons.value(course_id)) {
            const auto ldata = db_data->Nodes.value(lid);
            int len = ldata.value("LENGTH").toInt();
            // Add possible chosen room
            QStringList roomlist(rooms);
            auto fr{ldata.value("FLEXIBLE_ROOM")};
            if (fr.isUndefined()) {
                if (course.contains("ROOM_CHOICE"))
                    roomlist.append("?");
            } else
                roomlist.append(db_data->get_tag(fr.toInt()));
            int d0 = ldata.value("DAY").toInt();
            if (d0 == 0) {
//TODO: Collect unplaced lessons

            } else {
                int d = db_data->days.value(d0);
                int h = db_data->hours.value(ldata.value("HOUR").toInt());
                for (const auto &tf : tiles) {
                    Tile *t = new Tile(grid,
                                       QJsonObject{
                                           {"TEXT", subject},
                                           {"TL", teacher},
                                           {"TR", tf.groups.join(",")},
                                           {"BR", roomlist.join(",")},
                                           {"LENGTH", len},
                                           {"DIV0", tf.offset},
                                           {"DIVS", tf.fraction},
                                           {"NDIVS", tf.total},
                                       },
                                       lid);
                    grid->place_tile(t, d, h);
                }
            }
        }
    }
}
