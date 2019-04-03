# WEB+DB PRESS Vol.110 特集2「［速習］gRPC」

技術評論社刊「[WEB+DB PRESS Vol.110](https://gihyo.jp/magazine/wdpress/archive/2019/vol110)」特集2「［速習］gRPC」のタスク管理マイクロサービスのコードです。

## 動作環境

タスク管理マイクロサービスは次の環境で動作確認済みです。

- Go 1.12
- grpc-go 1.19.1
- protobuf 3.7.0
- protoc-gen-go 1.3.1
- Docker
  - Docker Desktop 2.0.0.3（Mac、Windows）
  - Docker CE 18.09.3 ＋ Docker Compose 1.23.2（Ubuntu）

## 起動方法

### .protoファイルをコンパイルする

```
$ protoc -I=proto --go_out=plugins=grpc,paths=source_relative:./proto proto/activity/activity.proto
$ protoc -I=proto --go_out=plugins=grpc,paths=source_relative:./proto proto/task/task.proto
$ protoc -I=proto --go_out=plugins=grpc,paths=source_relative:./proto proto/user/user.proto
$ protoc -I=proto --go_out=plugins=grpc,paths=source_relative:./proto proto/project/project.proto
```

### 各サービスをビルドする

```
$ docker-compose build
```

### 各サービスを起動する

```
$ docker-compose up
```

## ライセンス

サンプルコードはMITライセンスで配布しています。

http://opensource.org/licenses/mit-license.php

## お問い合わせ

不具合があった場合は[本誌Webページ](https://gihyo.jp/magazine/wdpress/archive/2019/vol110)よりお問い合わせをお願いいたします。

## ご注意

本サンプルコード、特集の内容に基づく運用結果について、著者、ソフトウェアの開発元および提供元、株式会社技術評論社は一切の責任を負いかねますので、あらかじめご了承ください。