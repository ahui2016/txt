// 采用受 Mithril 启发的基于 jQuery 实现的极简框架 https://github.com/ahui2016/mj.js
import { mjElement, mjComponent, m, cc, span } from "./mj.js";
import * as util from "./util.js";

const FormAlerts = util.CreateAlerts();
const footerElem = util.CreateFooter();

const NaviBar = cc("div", {
  children: [
    util.LinkElem("/", { text: "Home" }),
    span(" .. "),
    util.LinkElem("/public/sign-in.html", { text: "Sign-in" }),
    span(" .. Key and Password"),
  ],
});

const aboutPage = m("div").append(
  '本软件的安全措施分为 "主密码" 与 "日常操作密钥"。主密码的唯一用途是获取当前密钥或生成新密钥，',
  "其他操作如果要求输入密码，一律是指日常操作密钥（以下简称密钥）。"
);

const aboutSecretKey = m("div").append(
  m("h3").text("Secret Key").addClass("mb-0"),
  m("hr"),
  m("p").text(
    "输入主密码，点击 Get Key 按钮可获取当前密钥。" +
      "获取密钥后会出现 Generate 按钮，点击该按钮可生成新的密钥。" +
      "一旦生成新密钥，旧密钥就会作废。"
  )
);

const CurrentKeyArea = cc("div");
CurrentKeyArea.init = function (key: util.CurrentKey) {
  const keyStarts = dayjs.unix(key.Starts).format("YYYY-MM-DD");
  const keyExpires = dayjs.unix(key.Expires).format("YYYY-MM-DD");
  const self = CurrentKeyArea.elem();
  self.append(
    m("div").append(
      span("Current Key"),
      m("input").addClass("ml-2").val(key.Key).prop({ readonly: true })
    )
  );
  self.append(m("div").text(`生效日期: ${keyStarts}`));
  self.append(m("div").text(`有效期: ${key.MaxAge} (天)`));
  if (key.IsGood) {
    self.append(
      m("div").append(span("状态: "), span("有效").addClass("alert-success"))
    );
    self.append(
      m("div").addClass("form-text").text(`(该密钥将于 ${keyExpires} 自动作废)`)
    );
  } else {
    self.append(
      m("div").append(span("状态: "), span("已过期").addClass("alert-danger"))
    );
    self.append(
      m("div").addClass("form-text").text(`该密钥已于 ${keyExpires} 作废`)
    );
  }
};

// https://www.chromium.org/developers/design-documents/form-styles-that-chromium-understands/
const UsernameInput = cc("input", { attr: { autocomplete: "username" } });
const PwdInput = cc("input", { attr: { autocomplete: "current-password" } });
const GetKeyBtn = cc("button", { text: "Get key" });
const GenKeyBtn = cc("button", { text: "Generate" });

const Form = cc("form", {
  children: [
    m("label").text("Master Password").attr({ for: PwdInput.raw_id }),
    m("div").append(
      m(UsernameInput).hide(),
      m(PwdInput)
        .addClass("form-textinput form-textinput-fat")
        .attr({ type: "password" }),
      m(FormAlerts),
      m("div")
        .addClass("text-right")
        .append(
          m(GetKeyBtn).on("click", (event) => {
            event.preventDefault();
            const pwd = util.val(PwdInput);
            if (!pwd) {
              util.focus(PwdInput);
              return;
            }
            util.disable(GetKeyBtn);
            util.ajax(
              {
                method: "POST",
                url: "/auth/get-current-key",
                alerts: FormAlerts,
                body: { password: pwd },
              },
              // success
              (resp) => {
                const currentKey = resp as util.CurrentKey;
                CurrentKeyArea.show();
                CurrentKeyArea.init!(currentKey);
                GenKeyBtn.show();
                FormAlerts.clear();
              },
              // fail
              (_, errMsg) => {
                util.enable(GetKeyBtn);
                FormAlerts.insert("danger", errMsg);
                util.focus(PwdInput);
              }
            );
          }),
          m(GenKeyBtn)
            .addClass("ml-2")
            .on("click", (event) => {
              event.preventDefault();
              const pwd = util.val(PwdInput);
              if (!pwd) {
                util.focus(PwdInput);
                return;
              }
              util.ajax(
                {
                  method: "POST",
                  url: "/auth/gen-new-key",
                  alerts: FormAlerts,
                  buttonID: GenKeyBtn.id,
                  body: { password: pwd },
                },
                // success
                (resp) => {
                  const currentKey = resp as util.CurrentKey;
                  CurrentKeyArea.elem().html("");
                  CurrentKeyArea.init!(currentKey);
                  PwdInput.elem().val("");
                }
              );
            })
            .hide()
        )
    ),
  ],
});

const aboutPassword = m("div").append(
  m("h3").text("Change Master Password").addClass("mb-0"),
  m("hr"),
  m("p").text("可在此修改主密码。")
);

const CurrentPwd = util.create_input("password");
const NewPwd = util.create_input();
const SubmitBtn = cc("button", { text: "Change Password" });
const PwdAlerts = util.CreateAlerts();

const PwdForm = cc("form", {
  children: [
    util
      .create_item(CurrentPwd, "Current Password", "")
      .attr({ autocomplete: "current-password" }),
    util
      .create_item(NewPwd, "New Password", "")
      .attr({ autocomplete: "new-password" }),
    m(PwdAlerts),
    m(SubmitBtn).on("click", (event) => {
      event.preventDefault();
      const body = {
        oldpwd: util.val(CurrentPwd),
        newpwd: util.val(NewPwd),
      };
      if (!body.oldpwd || !body.newpwd) {
        PwdAlerts.insert("danger", "当前密码与新密码都必填");
        return;
      }
      util.ajax(
        {
          method: "POST",
          url: "/auth/change-pwd",
          alerts: PwdAlerts,
          buttonID: SubmitBtn.id,
          body: body,
        },
        () => {
          PwdAlerts.clear().insert("success", "已成功更改主密码。");
          CurrentPwd.elem().val("");
          NewPwd.elem().val("");
        }
      );
    }),
  ],
});

$("#root").append(
  m(NaviBar).addClass("my-5"),
  aboutPage.addClass("my-3"),
  aboutSecretKey,
  m(Form),
  m(CurrentKeyArea).addClass("mb-5").hide(),
  aboutPassword.addClass("mt-5"),
  m(PwdForm),
  footerElem
);

init();

function init() {
  $("title").text("Password .. txt-online");
  util.focus(PwdInput);
}
