#!/bin/bash

echo "🚀 运行碰撞检测演示程序..."
echo "================================"

cd collision_demo
go run main.go

echo ""
echo "✅ 演示完成！"
echo ""
echo "📝 代码说明："
echo "1. 实现了完整的3D射线-三角形相交检测"
echo "2. 使用Möller-Trumbore算法进行精确碰撞检测"
echo "3. 包含包围盒预检测优化"
echo "4. 演示了法术碰撞、移动碰撞和视线检测"
echo "5. 模拟了魔兽世界风格的地图和建筑系统"