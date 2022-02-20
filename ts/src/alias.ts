// 采用受 Mithril 启发的基于 jQuery 实现的极简框架 https://github.com/ahui2016/mj.js
import { mjElement, mjComponent, m, cc, span, appendToList } from "./mj.js";
import * as util from "./util.js";
import { CreateCopyComp, MsgItem, TxtMsg } from "./txtmsg-item.js";

var last_alias = "";

const Alerts = util.CreateAlerts();
const Loading = util.CreateLoading("center");
const TextForCopy = CreateCopyComp();

const NaviBar = cc("div", {
  classes: "my-5",
  children: [
    util.LinkElem("/public/index.html", { text: "Home" }),
    span(" .. "),
    util.LinkElem("/public/perm.html", { text: "Perm" }),
    span(" .. "),
    util.LinkElem("/public/temp.html", { text: "Temp" }),
    span(" .. Aliases (有别名的消息)"),
  ],
});

const MsgList = cc("div");

const MoreBtn = cc("button", { text: "More" });
const MoreBtnArea = cc("div", {
  children: [
    m(MoreBtn).on("click", (e) => {
      e.preventDefault();
      getMoreAlias();
    }),
  ],
});

$("#root").append(
  m(NaviBar),
  m(Loading).addClass("my-3"),
  m(MsgList).addClass("mt-3"),
  m(Alerts),
  m(MoreBtnArea).addClass("mt-5 text-center").hide(),
  m("div").text(".").addClass("Footer"),
  m(TextForCopy).hide()
);

init();

function init() {
  $("title").text("Aliases .. txt-online");
  getMoreAlias();
}

function getMoreAlias(): void {
  const body = {
    bucket: "alias-bucket",
    start: last_alias,
    limit: -1, // 小于等于零表示采用默认值
  };
  util.ajax(
    {
      method: "POST",
      url: "/api/get-more-items",
      alerts: Alerts,
      buttonID: MoreBtn.id,
      contentType: "json",
      body: body,
    },
    (resp) => {
      const items = resp as TxtMsg[];
      if (items && items.length > 0) {
        appendToList(MsgList, items.map(MsgItem));
        if (last_alias == "") {
          MoreBtnArea.show();
        }
        last_alias = items[items.length - 1].Alias;
      } else {
        if (last_alias == "") {
          Alerts.insert("info", "空空如也");
        } else {
          Alerts.insert("info", "没有更多了");
          MoreBtnArea.hide();
        }
      }
    },
    undefined,
    () => {
      Loading.hide();
    }
  );
}
