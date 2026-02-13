# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).


## [1.2.2] - 2026-02-13

### Performance 🚀
* **Structural Improvements:** Znaczące przyspieszenie operacji cyklu życia encji dzięki optymalizacji indeksowania w Arche.
    * **Create Entity:** Tworzenie encji z komponentami jest teraz szybsze o **36% do 46%**.
    * **Migrate Component:** Dodawanie kolejnych komponentów (migracja struktury) przyspieszyło o **38.22%**.
    * **Add Tag:** Operacja dodawania tagów jest teraz o **21.28%** wydajniejsza.
* **Known Regressions:** Zmiany architektoniczne wprowadziły niewielkie opóźnienia w `Remove Component` (+11.04%) oraz `Get Component` (+40.81%).

<details>
  <summary>Click to see full benchstat results (Apple M1 Max)</summary>

| Task | Before | After | Δ % |
| :--- | :--- | :--- | :--- |
| **Create Entity with 1 Component** | 38.42 ns | 21.31 ns | **-44.53%** |
| **Create Entity with 2 Components** | 38.56 ns | 23.29 ns | **-39.60%** |
| **Create Entity with 3 Components** | 46.03 ns | 24.87 ns | **-45.97%** |
| **Create Entity with 4 Components** | 42.01 ns | 26.84 ns | **-36.11%** |
| **Migrate Component** | 60.68 ns | 37.49 ns | **-38.22%** |
| **Add Tag** | 47.60 ns | 37.47 ns | **-21.28%** |
| **Remove Component** | 10.69 ns | 11.87 ns | +11.04% |
| **Get Component** | 5.462 ns | 7.691 ns | +40.81% |

</details>

## [1.2.1] - 2026-02-11

### Performance 🚀
* **Significant performance boost.** The overall execution time (geomean) decreased by **19.22%**.
* **Filtering optimizations:** Operations like `Filter100` are now over **50% faster** (e.g., `View0_Filter100` dropped from ~288ns to ~134ns).
* **View rendering:** Standard view operations (`ViewX_All`) saw performance improvements ranging from **12% to 15%**.
* **Values extraction:** Most `Values` operations improved by roughly **11% to 17%**.

<details>
  <summary>Click to see full benchstat results (Apple M1 Max)</summary>

#### View & Query Operations
| Task | Before | After | Δ % |
| :--- | :--- | :--- | :--- |
| **View0_All** | 346.0 ns | 339.2 ns | -1.95% |
| **View1_All** | 515.0 ns | 439.6 ns | -14.64% |
| **View2_All** | 567.0 ns | 482.9 ns | -14.82% |
| **View3_All** | 728.1 ns | 639.5 ns | -12.17% |
| **View3WithTag_All** | 692.2 ns | 600.7 ns | -13.22% |
| **View4_All** | 728.4 ns | 639.9 ns | -12.14% |
| **View5_All** | 729.9 ns | 638.4 ns | -12.53% |
| **View6_All** | 728.9 ns | 634.3 ns | -12.98% |
| **View7_All** | 727.8 ns | 634.1 ns | -12.88% |
| **View8_All** | 730.4 ns | 634.3 ns | -13.16% |
| **View0_Filter100** | 288.2 ns | 134.2 ns | **-53.44%** |
| **View3_Filter100** | 466.9 ns | 232.7 ns | **-50.17%** |
| **View1_Values** | 400.5 ns | 332.8 ns | -16.92% |
| **View2_Values** | 477.4 ns | 401.2 ns | -15.97% |
| **View3_Values** | 640.2 ns | 555.8 ns | -13.19% |
| **View4_Values** | 634.3 ns | 554.3 ns | -12.61% |
| **View5_Values** | 635.1 ns | 554.6 ns | -12.67% |
| **View6_Values** | 635.5 ns | 553.9 ns | -12.84% |
| **View7_Values** | 634.4 ns | 642.5 ns | ~ |
| **View8_Values** | 635.4 ns | 564.7 ns | -11.13% |
| **View3_FilterValues100** | 471.9 ns | 230.4 ns | **-51.17%** |
| **View2_Filter100** | - | 197.3 ns | n/a |
| **GEOMEAN** | **573.2 ns** | **445.4 ns** | **-19.22%** |

</details>
