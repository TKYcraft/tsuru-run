# ビルド用の環境
FROM golang:1.18-alpine as builder

# go getにgit、glibcが必要となるためmusl-dev、cgoの関係でgccを使っているようなのでadd
RUN apk update && apk --no-cache add git gcc musl-dev

# ディレクトリの作成
RUN mkdir -p /go_usr/app/src /go_usr/app/build /go_usr/app/pkg

# ワーキングディレクトリの設定
WORKDIR /go_usr/app

RUN go mod init TKYcraft/tsuru-run

# ワーキングディレクトリの設定
WORKDIR /go_usr/app/pkg

# ffmpegのダウンロードと解凍
RUN wget https://johnvansickle.com/ffmpeg/builds/ffmpeg-git-amd64-static.tar.xz && \
    tar xvf ffmpeg-git-amd64-static.tar.xz && \
    cp -rpf $(ls -d ffmpeg-git*static/) /go_usr/app/pkg/ffmpeg

# 使用しているパッケージのインストール
RUN --mount=type=cache,target=${GOPATH}/src \
    go get \
    github.com/bwmarrin/discordgo \
    github.com/bwmarrin/dgvoice \
    layeh.com/gopus

# ホストのファイルをコンテナの作業ディレクトリに移行
COPY ./app/src /go_usr/app/src

# ワーキングディレクトリの設定
WORKDIR /go_usr/app

# buildを超えてキャッシュを保持し、ファイルが変更されていなければキャッシュの利用ができる
RUN --mount=type=cache,target=/.cache/go-build \
    CGO_ENABLED=1 \
    # スタティックリンクにすることでバイナリのみで動作するように変更
    go build -ldflags '-s -w -extldflags=-static' -o build/main.build src/main.go

# 実行ファイルのみの環境
FROM gcr.io/distroless/static as dev

COPY ./app/file /go_usr/file

# ビルド用の環境から実行ファイルをコピー
COPY --from=builder /go_usr/app/pkg/ffmpeg/ffmpeg /usr/bin/ffmpeg
COPY --from=builder /go_usr/app/build /go_usr/app/build

# 実行
ENTRYPOINT  [ "/go_usr/app/build/main.build" ]
