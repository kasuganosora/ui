#!/bin/bash
# Compile GLSL shaders to SPIR-V.
# Requires glslc from the Vulkan SDK.
# Usage: ./compile.sh
#
# After compilation, run: go generate ../shaders.go

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"

glslc -fshader-stage=vertex "$SCRIPT_DIR/rect.vert.glsl" -o "$SCRIPT_DIR/rect.vert.spv"
glslc -fshader-stage=fragment "$SCRIPT_DIR/rect.frag.glsl" -o "$SCRIPT_DIR/rect.frag.spv"

echo "Shaders compiled successfully."
echo "Now run: go generate in the vulkan package directory."
