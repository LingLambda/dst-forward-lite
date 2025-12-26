## 简介

>:warning: 本项目由于 Lagrange.Core 不可用的原因已停止更新，目前无法正常使用

本项目是[dst-forward](https://github.com/LingLambda/dst-forward)的重构简化版

旨在简化用户使用体验，修复koishi端一些无法预料的bug

得益于go的轻量跨平台特性，本项目内存占用相较于依赖koishi的 [dst-forward](https://github.com/LingLambda/dst-forward) ，仅为运行整个koishi项目的1/20 （10-20MB），并且极其简单易用，适合新手，强烈建议使用此版本

## 使用方法：

- 在[relase](https://github.com/LingLambda/dst-forward-lite/releases)下载你的饥荒服务器所在的操作系统对应的二进制程序
- 运行一次程序并关闭，会自动生成 application.toml
- 在 application.toml 中配置你想绑定的群号，即你想要bot转发消息的群号 (其他配置可参照模板[application.tamplate.toml](https://github.com/LingLambda/dst-forward-lite/blob/master/application.template.toml))
- 运行程序并扫码登录
- 启动你的饥荒联机版服务器，安装并启用[此mod](https://steamcommunity.com/sharedfiles/filedetails/?id=3581042885)
- 在qq群或饥荒中任意发送一条消息查看效果

## 其他:

不管有没有问题都欢迎通过邮件联系我: [abc1514671906@163.com](mailto:abc1514671906@163.com)

## 感谢:

- [Lagrange.Core](https://github.com/LagrangeDev/Lagrange.Core)
- [LagrangeGo](https://github.com/LagrangeDev/LagrangeGo)
- [LagrangeGo-Template](https://github.com/ExquisiteCore/LagrangeGo-Template)
- 我自己
