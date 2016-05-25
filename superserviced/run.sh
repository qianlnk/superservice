rm ./logs/* -f
nohup sudo ./superserviced -log_dir=./logs/ -v=5 -logtostderr=false &
