go build main.go
rm -r pack-output
mkdir pack-output
cp ./main ./pack-output/youplus
cp -a ./pack/. ./pack-output