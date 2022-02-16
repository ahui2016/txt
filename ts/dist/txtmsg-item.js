// 采用受 Mithril 启发的基于 jQuery 实现的极简框架 https://github.com/ahui2016/mj.js
import { m, cc, span } from "./mj.js";
import * as util from "./util.js";
export function ItemID(id) {
    return "i" + id;
}
export function MsgItem(item) {
    const ItemAlerts = util.CreateAlerts();
    const self = cc('div', {
        id: ItemID(item.ID),
        classes: "TxtMsgItem",
        children: [
            m('div').addClass('text-grey').append(span(`[${indexOf(item)}] ${item.ID}`)),
            m('div').text(item.Msg),
            m(ItemAlerts),
        ],
    });
    return self;
}
function indexOf(item) {
    const prefix = item.Cat == "Category-Temporary" ? "T" : "P";
    return `${prefix}${item.Index}`;
}
