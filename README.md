Smartmeter
======================

ビルド
------

依存ライブラリ取得

    gb vendor restore

Raspberry PI の場合のみ

    cd vendor/src/golang.org/x/sys/unix
    chmod +x *.sh *.pl
    GOOS=linux GOARCH=arm ./mkall.sh

ビルド

    gb build


使い方
------

smartmeter.conf にBルートのIDとパスワードを設定します。

    [routeb]
    id = "00000000000000000000000000000000"
    password = "************"
    
    [database]
    host = "localhost"
    port = 8089
    
    [logger]
    level = "info"


    ./smartmeter -c smartmeter.conf
