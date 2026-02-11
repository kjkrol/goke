# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.2.1] - 2026-02-11

### Performance 🚀
* **Significant performance boost.** The overall execution time (geomean) decreased by **19.22%**.
* **Filtering optimizations:** Operations like `Filter100` are now over **50% faster** (e.g., `View0_Filter100` dropped from ~288ns to ~134ns).
* **View rendering:** Standard view operations (`ViewX_All`) saw performance improvements ranging from **12% to 15%**.
* **Values extraction:** Most `Values` operations improved by roughly **11% to 17%**.

<details>
  <summary>Click to see full benchstat results (Apple M1 Max)</summary>

  ```text
  goos: darwin
  goarch: arm64
  cpu: Apple M1 Max
                           │   old.txt   │               new.txt               │
                           │   sec/op    │   sec/op     vs base                │
  View0_All-10               346.0n ± 2%   339.2n ± 1%   -1.95% (p=0.000 n=10)
  View1_All-10               515.0n ± 1%   439.6n ± 1%  -14.64% (p=0.000 n=10)
  View2_All-10               567.0n ± 0%   482.9n ± 2%  -14.82% (p=0.000 n=10)
  View3_All-10               728.1n ± 2%   639.5n ± 1%  -12.17% (p=0.000 n=10)
  View3WithTag_All-10        692.2n ± 0%   600.7n ± 1%  -13.22% (p=0.000 n=10)
  View4_All-10               728.4n ± 1%   639.9n ± 3%  -12.14% (p=0.000 n=10)
  View5_All-10               729.9n ± 0%   638.4n ± 1%  -12.53% (p=0.000 n=10)
  View6_All-10               728.9n ± 1%   634.3n ± 1%  -12.98% (p=0.000 n=10)
  View7_All-10               727.8n ± 0%   634.1n ± 0%  -12.88% (p=0.000 n=10)
  View8_All-10               730.4n ± 1%   634.3n ± 0%  -13.16% (p=0.000 n=10)
  View0_Filter100-10         288.2n ± 1%   134.2n ± 1%  -53.44% (p=0.000 n=10)
  View3_Filter100-10         466.9n ± 1%   232.7n ± 2%  -50.17% (p=0.000 n=10)
  View1_Values-10            400.5n ± 1%   332.8n ± 1%  -16.92% (p=0.000 n=10)
  View2_Values-10            477.4n ± 1%   401.2n ± 2%  -15.97% (p=0.000 n=10)
  View3_Values-10            640.2n ± 0%   555.8n ± 0%  -13.19% (p=0.000 n=10)
  View4_Values-10            634.3n ± 0%   554.3n ± 1%  -12.61% (p=0.000 n=10)
  View5_Values-10            635.1n ± 1%   554.6n ± 1%  -12.67% (p=0.000 n=10)
  View6_Values-10            635.5n ± 1%   553.9n ± 1%  -12.84% (p=0.000 n=10)
  View7_Values-10            634.4n ± 1%   642.5n ± 2%       ~ (p=0.060 n=10)
  View8_Values-10            635.4n ± 0%   564.7n ± 2%  -11.13% (p=0.000 n=10)
  View3_FilterValues100-10   471.9n ± 3%   230.4n ± 2%  -51.17% (p=0.000 n=10)
  View2_Filter100-10                         197.3n ± 1%
  geomean                    573.2n        445.4n       -19.22%
  ```

</details>