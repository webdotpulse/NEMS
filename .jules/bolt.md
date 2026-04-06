## 2024-05-18 - Vue Array Reactivity Overhead

**Learning:** When storing large datasets fetched from APIs in Vue 3 (like the energy time-series data for charts), using `ref()` creates a deep reactive proxy for every object and nested property in the array. This creates noticeable reactivity overhead. Changing to `shallowRef()` skips the deep proxying while still triggering reactivity when the array is fully replaced (`energySeries.value = newData`), which is the standard pattern for API responses.

**Action:** Always use `shallowRef` for large arrays or complex objects that are replaced wholesale rather than mutated deeply to save CPU cycles and memory. Also, remember to check `package.json` or lockfiles (`package-lock.json` vs `pnpm-lock.yaml`) to identify the correct package manager before attempting to build the project.

## 2024-05-20 - Go High-Frequency Loop Pointer Allocations

**Learning:** In the `broadcastState()` function in `backend/poller.go`, which is called continuously by the 1-second polling loop, pointers were being reallocated inside the loop (e.g. `v := 0.0; totalGrid = &v`). Every iteration dynamically allocates a new pointer on the heap, leading to a build-up of unnecessary GC pressure.

**Action:** When tracking optional states or aggregating totals in hot loops, use simple local primitives (like `float64` and `bool`) to track values and state. Only allocate pointers at the very end of the function when assembling the final struct. This eliminates intermediate heap allocations.
