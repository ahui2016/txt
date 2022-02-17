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
  const self = cc("div", {
    id: ItemID(item.ID),
    classes: "TxtMsgItem",
    children: [
      m("div")
        .addClass("text-grey")
        .append(
          span(`[${indexOf(item)}] ${item.ID}`),
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
                        item.Cat == "Category-Temporary"
                          ? "永久消息"
                          : "暂存消息";
                      ItemAlerts.insert(
                        "success",
                        `已转换至[${after}], 3 秒后会自动刷新页面。`
                      );
                      setTimeout(() => {
                        location.reload();
                      }, 3000);
                    }
                  );
                }),
              util.LinkElem("#", { text: "edit" }).attr({ title: "修改/别名" }),
              util.LinkElem("#", { text: "del" }).attr({ title: "彻底删除" }),
              util.LinkElem("#", { text: "copy" }).attr({ title: "复制内容" })
            )
        ),
      m("div").text(item.Msg),
      m(ItemAlerts),
    ],
  });
  return self;
}

function indexOf(item: TxtMsg): string {
  const prefix = item.Cat == "Category-Temporary" ? "T" : "P";
  return `${prefix}${item.Index}`;
}
