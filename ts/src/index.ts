// 采用受 Mithril 启发的基于 jQuery 实现的极简框架 https://github.com/ahui2016/mj.js
import { mjElement, mjComponent, m, cc, span, appendToList } from "./mj.js";
import * as util from "./util.js";
import { CreateCopyComp, MsgItem, TxtMsg } from "./txtmsg-item.js";

const Alerts = util.CreateAlerts();
const Loading = util.CreateLoading("center");
const footerElem = util.CreateFooter();
const TextForCopy = CreateCopyComp();

const titleArea = m("div").addClass("text-center").append(m("h1").text("txt"));

const GotoSignIn = util.CreateGotoSignIn();

const MsgList = cc("div");

const MsgInput = util.create_textarea();
const SendBtn = cc("button", { text: "Send" });
const FormAlerts = util.CreateAlerts();

const Form = cc("form", {
  children: [
    m(MsgInput)
      .addClass("form-textinput form-textinput-fat")
      .attr({ placeholder: "New message" }),
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
            }
          );
        })
      ),
    m(FormAlerts),
  ],
});

$("#root").append(
  titleArea,
  m(Loading).addClass("my-3"),
  m(Alerts),
  m(GotoSignIn).addClass("my-3").hide(),
  m(Form).hide(),
  m(MsgList).addClass("mt-3"),
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
        Form.elem().show();
        util.focus(MsgInput);
        getRecent();
      } else {
        GotoSignIn.elem().show();
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
