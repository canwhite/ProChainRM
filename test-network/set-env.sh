#!/bin/bash
# Environment setup script for Hyperledger Fabric test network
# Usage: source set-env.sh

# 这个是标记命令行工具
export PATH=${PWD}/../bin:$PATH

# 简单说就是告诉 peer 命令"怎么工作"的配置文件路径。
export FABRIC_CFG_PATH=$PWD/../config/

# Environment variables for Org1
# 启用tls安全连接
export CORE_PEER_TLS_ENABLED=true
# 指定组织 MSP ID（成员服务提供者）
export CORE_PEER_LOCALMSPID=Org1MSP
# Org1 peer 的 TLS根证书路径，用于验证身份
export CORE_PEER_TLS_ROOTCERT_FILE=${PWD}/organizations/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/tls/ca.crt
# Org1 管理员的身份证书和私钥路径
export CORE_PEER_MSPCONFIGPATH=${PWD}/organizations/peerOrganizations/org1.example.com/users/Admin@org1.example.com/msp
# Org1 peer 节点的网络地址和端口
export CORE_PEER_ADDRESS=localhost:7051

echo "Environment variables set for Org1"
echo "Peer CLI is ready to use"