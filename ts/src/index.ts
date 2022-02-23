// 采用受 Mithril 启发的基于 jQuery 实现的极简框架 https://github.com/ahui2016/mj.js
import { mjElement, mjComponent, m, cc, span, appendToList } from "./mj.js";
import * as util from "./util.js";
import { CreateCopyComp, MsgItem, TxtMsg } from "./txtmsg-item.js";

const Alerts = util.CreateAlerts();
const Loading = util.CreateLoading("center");
const footerElem = util.CreateFooter();
const TextForCopy = CreateCopyComp();

const titleArea = m("div")
  .addClass("text-center")
  .append(m("h1").text("txt online"));

const GotoSignIn = util.CreateGotoSignIn();

const NaviBar = cc("div", {
  classes: "text-center",
  children: [
    util.LinkElem("/public/search.html", { text: "Search", title: "查找" }),
    span(" | "),
    util.LinkElem("/public/temp.html", { text: "Temp", title: "暂存消息" }),
    span(" | "),
    util.LinkElem("/public/perm.html", { text: "Perm", title: "永久消息" }),
    span(" | "),
    util.LinkElem("/public/alias.html", { text: "Alias", title: "别名" }),
    span(" | "),
    util.LinkElem("/public/config.html", { text: "Config", title: "设定" }),
  ],
});

const MsgList = cc("div");

const MsgInput = util.create_textarea();
const SendBtn = cc("button", { text: "Send" });
const FormAlerts = util.CreateAlerts();

const Form = cc("form", {
  children: [
    m(MsgInput)
      .addClass("form-textinput form-textinput-fat")
      .attr({ placeholder: "New message" }),
    m(FormAlerts),
    m("div")
      .addClass("text-right")
      .append(
        m(SendBtn).on("click", (e) => {
          e.preventDefault();
          util.ajax(
            {
              method: "POST",
              url: "/api/add",
              alerts: FormAlerts,
              buttonID: SendBtn.id,
              body: { msg: util.val(MsgInput, "trim") },
            },
            () => {
              FormAlerts.insert("success", "发送成功, 3 秒后会自动刷新页面。");
              setTimeout(() => {
                location.reload();
              }, 3000);
            },
            (_, errMsg) => {
              if (errMsg.includes("same as last")) {
                FormAlerts.insert(
                  "info",
                  "与最近一条暂存消息重复，不重复插入。"
                );
              } else {
                FormAlerts.insert("danger", errMsg);
              }
            }
          );
        })
      ),
  ],
});

$("#root").append(
  titleArea,
  m(NaviBar).addClass("my-5"),
  m(Loading).addClass("my-5"),
  m(Alerts),
  m(GotoSignIn).addClass("my-3").hide(),
  m(Form).hide(),
  m(MsgList).addClass("mb-5"),
  footerElem.hide(),
  m(TextForCopy).hide()
);

init();

function init() {
  checkSignIn();
}

function checkSignIn(): void {
  util.ajax(
    { method: "GET", url: "/auth/is-signed-in", alerts: Alerts },
    (resp) => {
      const yes = resp as boolean;
      if (yes) {
        Form.show();
        util.focus(MsgInput);
        getRecent();
      } else {
        GotoSignIn.show();
      }
    },
    undefined,
    () => {
      Loading.hide();
    }
  );
}

function getRecent(): void {
  util.ajax(
    { method: "GET", url: "/api/recent-items", alerts: Alerts },
    (resp) => {
      const items = resp as TxtMsg[];
      if (items && items.length > 0) {
        appendToList(MsgList, items.map(MsgItem));
        if (items.length >= 5) {
          footerElem.show();
        }
      }
    }
  );
}
