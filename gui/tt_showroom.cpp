#include "tt_showroom.h"
#include <QJsonArray>

ShowRoom::ShowRoom(TtGrid *grid, DBData *db_data, int room_id)
{
    QJsonValue r{room_id};
    // I need to go through all the courses and perhaps their lessons
    for (int cid : db_data->Tables.value("COURSES")) {
        const auto course = db_data->Nodes.value(cid);
        bool found = course.value("FIXED_ROOMS").toArray().contains(r);
        for (int lid : db_data->course_lessons.value(cid)) {
            const auto ldata = db_data->Nodes.value(lid);
            if (!found && ldata.value("FLEXIBLE_ROOM") != r) continue;
            QStringList teachers;
            auto tlist = course.value("TEACHERS").toArray();
            for (const auto & t : tlist) {
                teachers.append(db_data->get_tag(t.toInt()));
            }
            QStringList groups;
            auto glist = course.value("GROUPS").toArray();
            for (const auto & g : glist) {
                // Combine class and group
                auto node = db_data->Nodes.value(g.toInt());
                auto gtag = node.value("TAG").toString();
                auto ctag = db_data->get_tag(node.value("CLASS").toInt());
                if (gtag.isEmpty()) {
                    groups.append(ctag);
                } else {
                    groups.append(ctag + "." + gtag);
                }
            }
            QString subject = db_data->get_tag(course.value("SUBJECT").toInt());
            int d0 = ldata.value("DAY").toInt();
            if (d0 == 0) {
//TODO: Collect unplaced lessons

            } else {
                int d = db_data->days.value(d0);
                int h = db_data->hours.value(ldata.value("HOUR").toInt());
                Tile *t = new Tile(grid,
                    QJsonObject {
                       {"TEXT", groups.join(",")},
                       {"TL", subject},
                       {"BR", teachers.join(",")},
                       {"LENGTH", ldata.value("LENGTH").toInt()},
                       {"DIV0", 0},
                       {"DIVS", 1},
                       {"NDIVS", 1},
                    },
                    lid
                );
                grid->place_tile(t, d, h);
            }
        }
    }
}
