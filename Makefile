normal: git

git:
	@read -p "请输入提交说明: " msg; \
	if [ -z "$$msg" ]; then \
	  echo "提交说明不能为空，已取消提交。"; \
	  exit 1; \
	fi; \
	git config user.email "org-lib@163.com"; \
	git config user.name "org-lib"; \
	git pull; \
	git add .; \
	git commit -m "$$msg"; \
	git push

ReadMe:
	echo			APIHost := fmt.Sprintf("http://%v:%v@%v:%v/", epollo.Clickhouse_Username, epollo.Clickhouse_Password, epollo.Clickhouse_Host, epollo.Clickhouse_Port) // 根据实际情况设置
	echo			zap.L().Info("dbhouse.crontab", zap.String("Pullawslogs 启动解析aws慢日志:", *dbInstance.DBInstanceIdentifier))
	echo 			gogogo.ParseLogFile(fmt.Sprintf("%v", filePath), APIHost, _table, "mysql", []string{fmt.Sprintf("monitor_name=%v", *dbInstance.DBInstanceIdentifier), fmt.Sprintf("cluster_name=%v", _tmpClusterInfo.Clusterid)})
