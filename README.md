# 发个东西
一个局域网文字/文件P2P传输工具
> 项目中仅在线用户列表和WebRTC信令迫不得已需要一个轻量化的服务，其他数据传输都采用了基于WebRTC的点对点传输，不经过中间服务器，所以局域网内互传一些文字/文件都比较快。

demo演示：https://fagedongxi.com

## 优点
无需安装任何软件，打开浏览器，无需登录直接传输。

## 缺点
接收大文件比较吃内存（单文件几百兆一般没问题）

## 场景：
比如新装的win系统需要从mac系统传一些需要🪜才能下载的软件或者搜到的一些东西

## 服务端部署（仅部署服务端不行，一定看到最后的“网页部署”）：
部署介绍：https://v.douyin.com/iUWewPmf/

## GO 服务端

### 源码方式
1. 安装nodejs，node版本没有测试，我用的是 `16.20.2`
2. 下载源码（服务端仅需要`server`目录）
3. 进入 `server` 目录，运行 `npm install`
4. 运行 `npm run start [port]` ，例如 `npm run start 8081`

### 二进制方式
* 下载对应平台的可执行文件，直接执行即可（服务端）
* 默认监听 `8081` 端口，可通过参数指定端口，例如 `./internal-chat-linux 8082`
* 如果你用windows，可参考 https://v.douyin.com/CeiJahpLD/ 注册成服务

### 服务端nginx反向代理配置参考（可选）
> 服务端用反向代理的好处：可以直接用certbot申请https证书，然后直接用wss协议。
> 如果采用下方的配置反向代理，注意在客户端配置`wsUrl`变量的时候，需要加 `/ws`，否则不用
```
  location /ws/ {
    proxy_pass http://localhost:8081/;
    proxy_http_version 1.1;
    proxy_set_header Host $host;
    proxy_set_header Upgrade $http_upgrade;
    proxy_set_header Connection "upgrade";
    proxy_set_header X-Real-IP $remote_addr;
    proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
  }
```

## 网页部署：
1. 下载源码并修改`www/index.js`第一行代码`wsUrl`变量（如果服务端配置了反向代理，这里路径最后要加`/ws`，否则不用）
2. 直接将`www`用nginx部署成一个静态网站即可，具体配置参考 `nginxvhost.conf`。如果你没有域名，将 `server_name` 写成 `_` 即可（属于nginx基础知识）
3. 访问 `http://your.domain.com/` 即可

## 免责声明：
本项目仅用于学习交流，请勿用于非法用途，否则后果自负。
