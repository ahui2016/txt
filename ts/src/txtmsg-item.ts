// 采用受 Mithril 启发的基于 jQuery 实现的极简框架 https://github.com/ahui2016/mj.js
import { mjElement, mjComponent, m, cc, span, appendToList } from "./mj.js";
import * as util from "./util.js";

export interface TxtMsg {
  ID: string; // DateID, 既是 id 也是创建日期
  UserID: string; // 暂时不使用，以后升级为多用户系统时使用
  Alias: string; // 别名，要注意与 Alias bucket 联动。
  Msg: string; // 消息内容
  Cat: string; // 类型（比如暂存、永久）
  Index: number; // 流水号，每当插入或删除条目时，需要更新全部条目的流水号
}

export function ItemID(id: string): string {
  return "i" + id;
}

export function MsgItem(item: TxtMsg): mjComponent {
  const ItemAlerts = util.CreateAlerts();
  let title = `[${indexOf(item)}] ${item.ID}`;
  if (item.Alias) {
    title = `[${indexOf(item)}] [${item.Alias}] ${item.ID}`
  }
  const self = cc("div", {
    id: ItemID(item.ID),
    classes: "TxtMsgItem",
    children: [
      m("div")
        .addClass("text-grey")
        .append(
          span(title),
          m("div")
            .addClass("ItemButtons")
            .append(
              span("|").addClass("ml-2"),
              util
                .LinkElem("#", { text: "toggle" })
                .addClass("toggle-btn")
                .attr({ title: "暂存/永久" })
                .on("click", (e) => {
                  e.preventDefault();
                  const buttonID = self.id + " .toggle-btn";
                  util.ajax(
                    {
                      method: "POST",
                      url: "/api/toggle-category",
                      alerts: ItemAlerts,
                      buttonID: buttonID,
                      body: { id: item.ID },
                    },
                    () => {
                      const after =
                        item.Cat == "Temporary"
                          ? "永久消息"
                          : "暂存消息";
                      ItemAlerts.insert(
                        "success",
                        `已转换至[${after}], 3 秒后会自动刷新页面。`
                      );
                      reload();
                    }
                  );
                }),
              util
                .LinkElem("/public/edit.html?id=" + item.ID, {
                  text: "edit",
                  blank: true,
                })
                .attr({ title: "修改/别名" }),
              util
                .LinkElem("#", { text: "del" })
                .attr({ title: "彻底删除" })
                .addClass("del-btn")
                .on("click", (e) => {
                  e.preventDefault();
                  const buttonID = self.id + " .del-btn";
                  util.ajax(
                    {
                      method: "POST",
                      url: "/api/delete",
                      alerts: ItemAlerts,
                      buttonID: buttonID,
                      body: { id: item.ID },
                    },
                    () => {
                      self.elem().css({ "font-size": "small", color: "grey" });
                      self.elem().find(".ItemButtons").removeClass().hide();
                      ItemAlerts.insert(
                        "info",
                        "已删除。注意：全部消息的流水号已改变，请刷新页面获取新的流水号。"
                      );
                    }
                  );
                }),
              util
                .LinkElem("#", { text: "copy" })
                .attr({ title: "复制内容" })
                .on("click", (e) => {
                  e.preventDefault();
                  copyToClipboard(item.Msg);
                  ItemAlerts.insert("success", "复制成功");
                })
            )
        ),
      m("div").text(item.Msg),
      m(ItemAlerts),
    ],
  });
  return self;
}

function indexOf(item: TxtMsg): string {
  const prefix = item.Cat == "Temporary" ? "T" : "P";
  return `${prefix}${item.Index}`;
}

export function CreateCopyComp(): mjComponent {
  return cc("textarea", { id: "text-input-for-copy" });
}

function copyToClipboard(s: string): void {
  const textElem = $("#text-input-for-copy");
  textElem.show();
  textElem.val(s).trigger("select");
  document.execCommand("copy"); // execCommand 准备退役了，但仍没有替代方案，因此继续用。
  textElem.val("");
  textElem.hide();
}

function reload(second = 3): void {
  setTimeout(() => {
    location.reload();
  }, second * 1000);
}
