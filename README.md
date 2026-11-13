# Motty issue #712 — selected Windows GUI evidence

Screenshots captured during the Windows VM verification on 2026-11-13. These crops preserve the original pixels and omit unrelated desktop areas. The GUI test build used base `9fd5327` with the same five launcher-fix files recorded in the source manifest; the publication commit is rebased onto newer upstream main.

1. [Official Pi launcher discovery](windows-official-pi-auto-discovery.png): Windows 10 automatically detected the official Pi 1.5.0 `pi.cmd` in the add-provider dialog.

   官方 Pi 启动器发现：Windows 10 的添加服务商弹窗自动找到了官方安装的 Pi 1.5.0 `pi.cmd`。

2. [Official Pi model check](windows-official-pi-default-model.png): Automatic provider lookup remains enabled. Authentication shows “status unknown” with the environment-key/custom-provider hint; the local model check reports one enabled CLI default model.

   官方 Pi 模型检查：服务商保持自动查找；认证显示“状态未知”，并提示环境密钥或自定义服务商仍可使用。本地模型检查显示一个已启用的 CLI 默认模型。

3. [npm Pi custom-provider reply](windows-npm-pi-custom-provider-reply.png): npm Pi 1.5.1, using an explicit `.ps1` launcher in a path containing spaces, returned a real OpenRouter reply in Motty. The UI shows completion in 8.8 seconds and the `default` selector.

   npm Pi 自定义服务商回复：npm Pi 1.5.1 使用含空格路径中的显式 `.ps1` 启动器，通过 OpenRouter 在 Motty 中返回了真实回复。界面显示 8.8 秒完成，模型选择器为 `default`。

The model check shown here represents Pi's local CLI `default` configuration. The reply demonstrates one request through the configured custom provider.

这里的模型检查对应 Pi 本地 CLI 的 `default` 配置；回复截图证明了通过已配置自定义服务商完成的一次请求。
