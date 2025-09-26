#!/bin/bash

# AzerothCore 真实 Recast Navigation Go 演示构建脚本
# 基于真实的 Recast Navigation 库实现

echo "🏰 AzerothCore 真实 Recast Navigation 构建脚本"
echo "=============================================="

# 检查依赖
echo "📋 检查构建依赖..."

# 检查 Go 环境
if ! command -v go &> /dev/null; then
    echo "❌ Go 未安装，请先安装 Go 1.19+"
    exit 1
fi

echo "✅ Go 版本: $(go version)"

# 检查 Recast Navigation 库
RECAST_PATH="../deps/recastnavigation"
if [ ! -d "$RECAST_PATH" ]; then
    echo "❌ Recast Navigation 库未找到: $RECAST_PATH"
    echo "💡 请确保 AzerothCore 项目已正确初始化子模块"
    exit 1
fi

echo "✅ Recast Navigation 库路径: $RECAST_PATH"

# 检查编译器
if ! command -v g++ &> /dev/null; then
    echo "❌ g++ 编译器未找到，请安装 build-essential"
    exit 1
fi

echo "✅ C++ 编译器: $(g++ --version | head -n1)"

# 构建 Recast Navigation 库
echo ""
echo "🔨 构建 Recast Navigation 库..."

cd "$RECAST_PATH"

# 创建构建目录
if [ ! -d "build" ]; then
    mkdir build
fi

cd build

# 使用 CMake 构建
if command -v cmake &> /dev/null; then
    echo "📦 使用 CMake 构建..."
    cmake .. -DCMAKE_BUILD_TYPE=Release -DRECASTNAVIGATION_DEMO=OFF -DRECASTNAVIGATION_TESTS=OFF
    make -j$(nproc)
    
    # 检查库文件
    if [ -f "libRecast.a" ] && [ -f "libDetour.a" ]; then
        echo "✅ Recast Navigation 库构建成功"
    else
        echo "❌ Recast Navigation 库构建失败"
        exit 1
    fi
else
    echo "⚠️  CMake 未找到，尝试手动编译..."
    
    # 手动编译 Recast
    cd ../Recast/Source
    g++ -c -O3 -fPIC *.cpp -I../Include
    ar rcs libRecast.a *.o
    rm *.o
    
    # 手动编译 Detour
    cd ../../Detour/Source
    g++ -c -O3 -fPIC *.cpp -I../Include -I../../Recast/Include
    ar rcs libDetour.a *.o
    rm *.o
    
    # 移动库文件
    mkdir -p ../../build
    mv ../Source/libDetour.a ../../build/
    mv ../../Recast/Source/libRecast.a ../../build/
    
    echo "✅ 手动编译完成"
fi

# 返回演示目录
cd - > /dev/null
cd ../go-recast-demo

# 设置环境变量
export CGO_CFLAGS="-I../deps/recastnavigation/Recast/Include -I../deps/recastnavigation/Detour/Include"
export CGO_LDFLAGS="-L../deps/recastnavigation/build -lRecast -lDetour -lstdc++ -lm"

echo ""
echo "🚀 构建 Go 演示程序..."

# 构建演示程序
if go build -o real_azerothcore_demo real_azerothcore_demo.go; then
    echo "✅ Go 演示程序构建成功"
else
    echo "❌ Go 演示程序构建失败"
    echo ""
    echo "🔧 故障排除建议:"
    echo "1. 检查 CGO 环境变量设置"
    echo "2. 确保 Recast Navigation 库已正确编译"
    echo "3. 检查头文件路径是否正确"
    echo "4. 尝试手动设置 LD_LIBRARY_PATH"
    exit 1
fi

echo ""
echo "🎯 运行演示程序..."
echo "================================"

# 运行演示程序
if [ -f "./real_azerothcore_demo" ]; then
    ./real_azerothcore_demo
else
    echo "❌ 演示程序未找到"
    exit 1
fi

echo ""
echo "✅ 演示完成!"
echo ""
echo "📚 更多信息:"
echo "   - 真实演示程序: real_azerothcore_demo.go"
echo "   - AzerothCore 项目: https://github.com/azerothcore/azerothcore-wotlk"
echo "   - Recast Navigation: https://github.com/recastnavigation/recastnavigation"