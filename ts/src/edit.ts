// 采用受 Mithril 启发的基于 jQuery 实现的极简框架 https://github.com/ahui2016/mj.js
import { mjElement, mjComponent, m, cc, span } from "./mj.js";
import { TxtMsg } from "./txtmsg-item.js";
import * as util from "./util.js";

const id = util.getUrlParam("id");
let tm: TxtMsg;

const Alerts = util.CreateAlerts();
const Loading = util.CreateLoading("center");
const footerElem = util.CreateFooter();

const NaviBar = cc("div", {
  classes: "my-5",
  children: [util.LinkElem("/", { text: "home" }), span(" .. Edit")],
});

const ID_Input = util.create_input();
const CatInput = util.create_input();
const AliasInput = util.create_input();
const MsgInput = util.create_textarea(5);
const FormAlerts = util.CreateAlerts();
const HiddenBtn = cc("button", { id: "submit", text: "submit" }); // 这个按钮是隐藏不用的，为了防止按回车键提交表单
const SubmitBtn = cc("button", { text: "Submit" });

const Form = cc("form", {
  children: [
    util.create_item(ID_Input, "ID", ""),
    util.create_item(
      CatInput,
      "Category",
      "类型（暂存/永久），可在消息列表中点击 toggle 按钮转换类型。"
    ),
    util.create_item(
      AliasInput,
      "Alias",
      "别名(可留空)。用于方便命令行精准指定消息。"
    ),
    util.create_item(MsgInput, "Message", "文本消息内容(必填)"),
    m(HiddenBtn)
      .hide()
      .on("click", (e) => {
        e.preventDefault();
        return false; // 这个按钮是隐藏不用的，为了防止按回车键提交表单。
      }),
    m(SubmitBtn).on("click", (e) => {
      e.preventDefault();
      const msg = util.val(MsgInput, "trim");
      if (!msg) {
        FormAlerts.insert("danger", "Message(文本消息内容)必填");
        util.focus(MsgInput);
        return;
      }
      const body = {
        id: id,
        alias: util.val(AliasInput, "trim"),
        msg: msg,
      };
      if (body.alias == tm.Alias && body.msg == tm.Msg) {
        Alerts.clear().insert("success", "修改成功(内容无变化)");
        Form.hide();
        return;
      }
      util.ajax(
        {
          method: "POST",
          url: "/api/edit",
          alerts: FormAlerts,
          buttonID: SubmitBtn.id,
          body: body,
        },
        () => {
          Alerts.clear().insert("success", "修改成功");
          Form.hide();
        }
      );
    }),
    m(FormAlerts),
  ],
});

$("#root").append(
  m(NaviBar).addClass("my-3"),
  m(Loading).addClass("my-3"),
  m(Alerts).addClass("my-3"),
  m(Form).hide(),
  m("div").text(".").addClass("Footer")
);

init();

function init() {
  $("title").text("Edit .. txt");
  if (!id) {
    Loading.hide();
    Alerts.insert("danger", "未指定 id");
    return;
  }
  loadData();
}

function loadData(): void {
  util.ajax(
    { method: "POST", url: "/api/get-by-id", alerts: Alerts, body: { id: id } },
    (resp) => {
      tm = resp as TxtMsg;
      Form.show();
      ID_Input.elem().val(tm.ID);
      util.disable(ID_Input);
      CatInput.elem().val(tm.Cat);
      util.disable(CatInput);
      AliasInput.elem().val(tm.Alias);
      MsgInput.elem().val(tm.Msg);
    },
    undefined,
    () => {
      Loading.hide();
      util.focus(MsgInput);
    }
  );
}
