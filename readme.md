===
ckproxy针对公有云接口的代理功能；可根据配置接口描述, 代理公有云的各种接口

===
### 打包

linux:
    
    make pkg    
    

### 安装

linux：
    
    mkdir -p  /home/s/safe/
    tar -zxvf ckproxy.${VERSION}.tar.gz -C /home/s/safe/
    /home/s/safe/ckproxy/sbin/srvctl init  
    /home/s/safe/ckproxy/sbin/srvctl start
