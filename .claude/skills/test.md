# /test — GoUI Full Test Suite with Visual Regression

Run the full GoUI test suite including unit tests, integration tests, and visual
regression tests. After visual tests, analyze screenshots for rendering issues
and enter a TDD fix loop if problems are found.

## Workflow

### Phase 1: Unit & Integration Tests

1. Run the full unit test suite:
   ```
   cd d:/code/ui && go test ./math/... ./event/... ./core/... ./layout/... ./font/... ./widget/... ./render/capture/... -count=1 2>&1
   ```
2. Run platform-specific integration tests:
   ```
   cd d:/code/ui && go test ./platform/win32/... -count=1 -timeout=30s 2>&1
   ```
3. Run Vulkan backend tests:
   ```
   cd d:/code/ui && go test ./render/vulkan/... -count=1 -timeout=60s 2>&1
   ```

If any tests fail, stop here, diagnose the failure, fix the code, and re-run.
Do NOT proceed to visual tests until all unit/integration tests pass.

### Phase 2: Visual Regression Tests

4. Run visual tests which render real UI and capture screenshots:
   ```
   cd d:/code/ui && go test ./cmd/demo/... -run TestVisual -v -count=1 -timeout=120s 2>&1
   ```

   The visual test suite includes:
   - **TestVisualDemoFullUI** — Full demo layout, tree structure validation, distinct color count
   - **TestVisualButtonRendering** — All button variants visible with multi-color regions
   - **TestVisualInputRendering** — Input fields rendered with borders/placeholders
   - **TestVisualGridColors** — Grid columns show distinct graduated blue shades
   - **TestVisualMessageLoop** — No deadlock under 10-frame render loop
   - **TestVisualHitTestConsistency** — Every visible element with bounds is hit-testable
   - **TestVisualCommandBufferCoverage** — Draw() generates proportional commands to visible elements
   - **TestVisualFrameConsistency** — Consecutive frames produce identical output

5. Screenshots are saved to `cmd/demo/testdata/screenshots/`.
   Read each PNG file using the Read tool to **visually inspect** them:
   ```
   Read: d:/code/ui/cmd/demo/testdata/screenshots/demo_full_ui.png
   Read: d:/code/ui/cmd/demo/testdata/screenshots/button_rendering.png
   Read: d:/code/ui/cmd/demo/testdata/screenshots/input_rendering.png
   Read: d:/code/ui/cmd/demo/testdata/screenshots/grid_colors.png
   Read: d:/code/ui/cmd/demo/testdata/screenshots/message_loop.png
   ```

### Phase 3: Screenshot Analysis

6. For each screenshot, analyze using both **pixel data** and **tree structure**:

   **Tree-level checks** (automated by tests):
   - `dumpTree()` — Print full element tree with bounds, visibility, text content
   - `verifyAllVisibleHaveBounds()` — Every visible element has non-zero layout bounds
   - `verifyNoOverlappingSiblings()` — No sibling elements at identical positions
   - `countDistinctColors()` — Region has expected visual complexity
   - HitTest center of every element — All bounded elements are hittable

   **Visual checks** (human/AI review of screenshots):
   - **Black screen** — Rendering pipeline broken (check shader, vertex buffer, swapchain)
   - **Uniform color** — Widgets not drawing (check Draw() methods, layout bounds)
   - **Missing regions** — Layout not reaching all elements (check layout recursion depth)
   - **Overlapping elements** — Layout incorrect (check sibling positioning)
   - **Garbled pixels** — Vertex data or shader issue (check buffer offsets, NDC conversion)
   - **Wrong colors** — SRGB gamma, BGRA channel swap (check ReadPixels conversion)
   - **Black borders** — Backend width/height mismatch with swapchain extent (check DPI handling)
   - **Frozen window** — Missing runtime.LockOSThread (check platform init)

### Phase 4: TDD Fix Loop (if issues found)

7. If any visual issue is detected:
   a. **Identify** — Describe the specific rendering problem from screenshot + tree dump
   b. **Write a failing test** — Add or tighten an assertion that captures the bug:
      - Pixel color at specific coordinates: `verifyRegionHasColor()`
      - Distinct color count in region: `countDistinctColors()`
      - Element bounds validation: `verifyAllVisibleHaveBounds()`
      - Tree structure: `dumpTree()` + manual inspection
   c. **Fix the code** — Make the minimal change to fix the issue
   d. **Re-run tests**:
      ```
      cd d:/code/ui && go test ./... -count=1 -timeout=120s 2>&1
      ```
   e. **Re-capture screenshots**:
      ```
      cd d:/code/ui && go test ./cmd/demo/... -run TestVisual -v -count=1 -timeout=120s 2>&1
      ```
   f. **Re-analyze** — Read new screenshots with the Read tool
   g. **Loop** — Repeat until all tests pass and screenshots look correct

8. When complete, report:
   - Summary of tests run and results
   - Any fixes made during the TDD loop
   - Known visual limitations (e.g., placeholder text rects)

## Infrastructure Capabilities

The test harness leverages these existing capabilities:

| Capability | Package | Usage |
|---|---|---|
| Screenshot capture | `render/capture` | `Screenshot(backend)` → `*image.RGBA` |
| Golden comparison | `render/capture` | `MustMatchGolden(t, backend, path, threshold)` |
| Image diff | `render/capture` | `Compare(a, b, threshold)` → `DiffResult` |
| PSNR metric | `render/capture` | `PSNR(a, b)` → dB value |
| Save/Load PNG | `render/capture` | `SavePNG(img, path)` / `LoadPNG(path)` |
| Tree traversal | `core.Tree` | `Walk(root, fn)` — depth-first traversal |
| Hit testing | `core.Tree` | `HitTest(x, y)` → `ElementID` |
| Element inspection | `core.Element` | `.Type()`, `.Layout().Bounds`, `.TextContent()`, `.IsVisible()` |
| Element count | `core.Tree` | `ElementCount()` |
| Command buffer | `render.CommandBuffer` | `.Len()` — count of render commands |

## Key Files

- **Visual test harness**: `cmd/demo/visual_test.go`
- **Demo UI builder**: `cmd/demo/main.go` (buildUI, computeLayout)
- **Screenshot utilities**: `render/capture/capture.go`, `render/capture/testing.go`
- **Vulkan backend**: `render/vulkan/backend.go` (ReadPixels at line ~840)
- **Widget implementations**: `widget/*.go`
- **Win32 platform**: `platform/win32/platform.go`
- **Element tree**: `core/element.go` (Walk, HitTest, Get)

## Known Limitations

- Text is rendered as colored placeholder rectangles (font system not yet integrated)
- Manual layout in demo only; full layout engine integration pending
- Visual tests require GPU with Vulkan support
- Window created with `Visible=false` for headless testing
