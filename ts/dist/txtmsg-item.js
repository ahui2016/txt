// 采用受 Mithril 启发的基于 jQuery 实现的极简框架 https://github.com/ahui2016/mj.js
import { m, cc, span } from "./mj.js";
import * as util from "./util.js";
export function ItemID(id) {
    return "i" + id;
}
export function MsgItem(item) {
    const ItemAlerts = util.CreateAlerts();
    let title = `[${indexOf(item)}] ${item.ID}`;
    if (item.Alias) {
        title = `[${indexOf(item)}] [${item.Alias}] ${item.ID}`;
    }
    const self = cc("div", {
        id: ItemID(item.ID),
        classes: "TxtMsgItem",
        children: [
            m("div")
                .addClass("text-grey")
                .append(span(title), m("div")
                .addClass("ItemButtons")
                .append(span("|").addClass("ml-2"), util
                .LinkElem("#", { text: "toggle" })
                .addClass("toggle-btn")
                .attr({ title: "暂存/永久" })
                .on("click", (e) => {
                e.preventDefault();
                const buttonID = self.id + " .toggle-btn";
                util.ajax({
                    method: "POST",
                    url: "/api/toggle-category",
                    alerts: ItemAlerts,
                    buttonID: buttonID,
                    body: { id: item.ID },
                }, () => {
                    const after = item.Cat == "Category-Temporary"
                        ? "永久消息"
                        : "暂存消息";
                    ItemAlerts.insert("success", `已转换至[${after}], 3 秒后会自动刷新页面。`);
                    reload();
                });
            }), util
                .LinkElem("/public/edit.html?id=" + item.ID, {
                text: "edit",
                blank: true,
            })
                .attr({ title: "修改/别名" }), util
                .LinkElem("#", { text: "del" })
                .attr({ title: "彻底删除" })
                .addClass("del-btn")
                .on("click", (e) => {
                e.preventDefault();
                const buttonID = self.id + " .del-btn";
                util.ajax({
                    method: "POST",
                    url: "/api/delete",
                    alerts: ItemAlerts,
                    buttonID: buttonID,
                    body: { id: item.ID },
                }, () => {
                    self.elem().css({ "font-size": "small", color: "grey" });
                    self.elem().find(".ItemButtons").removeClass().hide();
                    ItemAlerts.insert("info", "已删除。注意：全部消息的顺序号已改变，请刷新页面获取新的顺序号。");
                });
            }), util
                .LinkElem("#", { text: "copy" })
                .attr({ title: "复制内容" })
                .on("click", (e) => {
                e.preventDefault();
                copyToClipboard(item.Msg);
                ItemAlerts.insert("success", "复制成功");
            }))),
            m("div").text(item.Msg),
            m(ItemAlerts),
        ],
    });
    return self;
}
function indexOf(item) {
    const prefix = item.Cat == "Category-Temporary" ? "T" : "P";
    return `${prefix}${item.Index}`;
}
export function CreateCopyComp() {
    return cc("textarea", { id: "text-input-for-copy" });
}
function copyToClipboard(s) {
    const textElem = $("#text-input-for-copy");
    textElem.show();
    textElem.val(s).trigger("select");
    document.execCommand("copy"); // execCommand 准备退役了，但仍没有替代方案，因此继续用。
    textElem.val("");
    textElem.hide();
}
function reload(second = 3) {
    setTimeout(() => {
        location.reload();
    }, second * 1000);
}
