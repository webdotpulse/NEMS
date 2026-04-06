## 2024-05-18 - Vue Array Reactivity Overhead

**Learning:** When storing large datasets fetched from APIs in Vue 3 (like the energy time-series data for charts), using `ref()` creates a deep reactive proxy for every object and nested property in the array. This creates noticeable reactivity overhead. Changing to `shallowRef()` skips the deep proxying while still triggering reactivity when the array is fully replaced (`energySeries.value = newData`), which is the standard pattern for API responses.

**Action:** Always use `shallowRef` for large arrays or complex objects that are replaced wholesale rather than mutated deeply to save CPU cycles and memory. Also, remember to check `package.json` or lockfiles (`package-lock.json` vs `pnpm-lock.yaml`) to identify the correct package manager before attempting to build the project.
