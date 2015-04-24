Deployment-Engine Project

rewrited by zhoujiawei,lidawei

Deployment Engine 特性

快速自动化部署一个完整的cloudfoundry集群，各组件并发安装

自动化部署单个cloudfoundry组件

可离线化的内网部署。各组件节点内核版本为3.2.0-61-virtual，即可离线，只需要Deployment Engine控制机联网
物理机、虚拟机或容器皆可部署，剥离对Iaas层的依赖
支持高可用部署，无单点
方便管理和运维集群的所有虚拟机节点
对各虚拟机节点实时监控，包括各节点的硬件配置（虚拟机核数vCPU、内存RAM、硬盘DISK）、资源使用量（CPU使用率、内存占用率、磁盘使用率）、组件运行状况
提供集群运维日志的显示与下载
提供各节点运行日志的下载
兼容cf-release版本：v160、v170、v194

Deployment Engine 节点系统要求

操作系统：Ubuntu10.04或Ubuntu12.04 64位
Ruby版本：Ruby 1.9.3p448
Rubygem版本：1.8.17
Yaml版本：0.1.4
监控软件：monit-5.2.4
vCPU核数：1个以上
内存大小：1024MB以上
磁盘大小：5120MB以上

Deployment Engine 安装与激活
Deployment Engine安装
Step 1：解压Deployment-Engine.tgz文件到任意目录下
Step 2：解压文件后，运行./local_build.sh进行Deployment Engine本地初始化，主要完成以下工作：
架设linux本地源：
    部署Apache服务器
    使缓存安装包可访问
    修改本地源
构建本地运行时环境：
    从linux本地源安装linux依赖
    安装ruby、rubygems等文件
    安装缓存的ruby gems包
Deployment Engine激活
初次使用Deployment Engine需要输入购买的序列号进行激活操作。执行./DeploymentEngineServer，输入产品序列号，确保Deployment Engine所在的这台部署机联网，否则无法进行激活或者部署集群。
！注意：不要随意修改Deployment Engine下的文件，文件损坏可能会导致Deployment Engine无法工作；不要把Deployment Engine这台部署机复制到别的VLAN中进行集群部署，否则会导致两台部署机都无法工作。
