# slg游戏服务器demo

### 概要
- 1.mysql数据落地，orm映射
- 2.事件处理支持中间件
- 3.服务器与服务器之间websocket连接
- 4.服务器与服务器之间rpc调用
- 5.高并发

### 多进程服务
- 1.httpserver  提供一些api调用
- 2.gateserver  网关，可以部署多个进行负债均衡，客户端的所有loginserver、chatserver、slgserver的消息都通过该服进行转发
- 3.loginserver 登录服，可以部署多个进行负债均衡
- 4.chatserver  聊天服，可以部署多个，原则上一个slgserver对应一个chatserver
- 5.slgserver   游戏服，可以部署多个，不同服之间的玩家数据不共通


# 客户端截图（因图片素材非自主生产，所以暂未公布仓库）
### 队伍征兵
![队伍征兵](https://s1.imagehub.cc/images/2021/04/23/d56cd91ba46b9ffd7b097dc4cb07bf5a.png)

### 占领领地
![占领领地](https://s1.imagehub.cc/images/2021/04/23/6e75b931ec76e840720c43f1a915eb85.png)

### 出征返回
![出征返回](https://s1.imagehub.cc/images/2021/04/23/2c6881d4caeff95de2d75c497ea0035e.png)

### 城内设施
![城内设施](https://s1.imagehub.cc/images/2021/04/23/6e99130a9fd3a104fa3c1177bc1b0947.png)

### 武将
![武将](https://s1.imagehub.cc/images/2021/04/23/a4cff5540d6a40a446b77fbfa58d8112.png)

### 武将详情
![武将详情](https://s1.imagehub.cc/images/2021/04/23/f579d57ae695b0686827e78ca3003340.png)

### 友方主城
![友方主城](https://s1.imagehub.cc/images/2021/04/23/1405cfa404e73b9bf.png)

### 敌方主城
![敌方主城](https://s1.imagehub.cc/images/2021/04/23/23b58a83e2baaf0f1.png)

### 军队前往敌方主城
![军队前往敌方主城](https://s1.imagehub.cc/images/2021/04/23/3.png)

### 抽卡结果
![抽卡结果](https://s1.imagehub.cc/images/2021/04/23/33ce47f51109b5b6f7d7370d1669878a.png)

### 战报
![战报](https://s1.imagehub.cc/images/2021/04/23/32e6bd48f6492e332640fdcd850a8118.png)

### 技能
![技能](https://s1.imagehub.cc/images/2021/04/23/e1417839fe85f2ec30fd1e9a07cfb61f.png)

### 联盟
![联盟](https://s1.imagehub.cc/images/2021/04/23/7d8c5e1b4119128673101a03a0ec1a8d.png)

### 聊天
![聊天](https://s1.imagehub.cc/images/2021/04/23/5c5785ceab3b9d4707bcb75548c570a0.png)
