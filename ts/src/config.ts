// 采用受 Mithril 启发的基于 jQuery 实现的极简框架 https://github.com/ahui2016/mj.js
import { mjElement, mjComponent, m, cc, span, appendToList } from "./mj.js";
import * as util from "./util.js";

const Alerts = util.CreateAlerts();
const Loading = util.CreateLoading("center");

const GotoSignIn = util.CreateGotoSignIn();

const NaviBar = cc("div", {
  classes: "my-5",
  children: [
    util.LinkElem("/", { text: "Home" }),
    span(" .. "),
    util.LinkElem("/public/sign-in.html", {
      text: "Sign-in/out",
      title: "登入/登出",
    }),
    span(" .. Config"),
  ],
});

const MaxAgeInput = util.create_input();
const MsgSizeInput = util.create_input();
const TempLimitInput = util.create_input();
const PageLimitInput = util.create_input();
const TimezoneInput = util.create_input();
const FormAlerts = util.CreateAlerts();
const HiddenBtn = cc("button", { id: "submit", text: "submit" }); // 这个按钮是隐藏不用的，为了防止按回车键提交表单
const SubmitBtn = cc("button", { text: "Submit" });

const Form = cc("form", {
  children: [
    util.create_item(
      MaxAgeInput,
      "Key Max Age",
      "密钥有效期（单位：天），不可小于 1 天"
    ),
    util.create_item(
      MsgSizeInput,
      "Message Size Limit",
      "每条消息的长度上限 (单位: byte), 不可小于 256。"
    ),
    util.create_item(
      TempLimitInput,
      "Temporary Messages Limit",
      "暂存消息条数上限，超过上限会自动删除旧消息。不可小于 1。"
    ),
    util.create_item(
      PageLimitInput,
      "Every Page Limit",
      "每页最多列出多少条消息，不可小于 1。"
    ),
    util.create_item(
      TimezoneInput,
      "Timezone Offset",
      '时区（例如 "+8" 表示北京时间, "-5" 表示纽约时间）, 不可频繁更改时区。'
    ),
    m(FormAlerts),
    m(HiddenBtn)
      .hide()
      .on("click", (e) => {
        e.preventDefault();
        return false; // 这个按钮是隐藏不用的，为了防止按回车键提交表单。
      }),
    m(SubmitBtn).on("click", (e) => {
      e.preventDefault();
      const body: util.ConfigForm = {
        KeyMaxAge: util.getNumber(MaxAgeInput),
        MsgSizeLimit: util.getNumber(MsgSizeInput),
        TempLimit: util.getNumber(TempLimitInput),
        EveryPageLimit: util.getNumber(PageLimitInput),
        TimeOffset: util.val(TimezoneInput),
      };
      util.ajax(
        {
          method: "POST",
          url: "/api/update-config",
          alerts: FormAlerts,
          buttonID: SubmitBtn.id,
          contentType: "json",
          body: body,
        },
        (resp) => {
          const warning = (resp as util.Text).message;
          Form.hide();
          Alerts.clear().insert("success", "更新成功");
          if (warning) {
            Alerts.insert("info", warning);
          }
        }
      );
    }),
  ],
});

$("#root").append(
  m(NaviBar).addClass("my-3"),
  m(Loading).addClass("my-3"),
  m(Alerts).addClass("my-3"),
  m(GotoSignIn).addClass("my-3").hide(),
  m(Form).hide(),
  m("div").text(".").addClass("Footer")
);

init();

function init() {
  $("title").text("Config .. txt");
  loadData();
}

function loadData() {
  util.ajax(
    { method: "GET", url: "/api/get-config", alerts: Alerts },
    (resp) => {
      const config = resp as util.ConfigForm;
      Form.show();
      MaxAgeInput.elem().val(config.KeyMaxAge);
      MsgSizeInput.elem().val(config.MsgSizeLimit);
      TempLimitInput.elem().val(config.TempLimit);
      PageLimitInput.elem().val(config.EveryPageLimit);
      TimezoneInput.elem().val(config.TimeOffset);
    },
    (that, errMsg) => {
      if (that.status == 401) {
        GotoSignIn.show();
      }
      Alerts.insert("danger", errMsg);
    },
    () => {
      Loading.hide();
    }
  );
}
