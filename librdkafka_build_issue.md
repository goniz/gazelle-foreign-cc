# librdkafka Build Failure: OpenSSL Headers Missing with SASL_OAUTHBEARER Enabled

## Summary

The librdkafka example fails to build due to a configuration conflict. The build has SSL disabled (`WITH_SSL OFF`) but SASL OAuth Bearer enabled (`WITH_SASL_OAUTHBEARER ON`), which requires OpenSSL headers.

## Steps to Reproduce

1. Navigate to the examples directory
```bash
cd examples
```

2. Run gazelle to update build files
```bash
bazel run :gazelle
```

3. Attempt to build librdkafka (note: the original command had typos)
```bash
# Original command (with typos):
bazel build //thirdpart/librdkafk:all

# Corrected command:
bazel build //thirdparty/librdkafka:all
```

## Expected Behavior

The librdkafka library should build successfully.

## Actual Behavior

The build fails with the following error:

```
ERROR: /workspace/examples/thirdparty/librdkafka/BUILD.bazel:29:11: Compiling src/rdkafka_sasl_oauthbearer.c failed: (Exit 1): gcc failed: error executing CppCompile command (from target //thirdparty/librdkafka:rdkafka) /usr/bin/gcc -U_FORTIFY_SOURCE -fstack-protector -Wall -Wunused-but-set-parameter -Wno-free-nonheap-object -fno-omit-frame-pointer -MD -MF ... (remaining 22 arguments skipped)

Use --sandbox_debug to see verbose messages from the sandbox and retain the sandbox build root for debugging
external/+_repo_rules+librdkafka/src/rdkafka_sasl_oauthbearer.c:36:10: fatal error: openssl/evp.h: No such file or directory
   36 | #include <openssl/evp.h>
      |          ^~~~~~~~~~~~~~~
compilation terminated.
```

## Root Cause Analysis

The issue stems from conflicting configuration directives in `examples/thirdparty/librdkafka/BUILD.bazel`:

```starlark
# gazelle:cmake_define WITH_SSL OFF
# gazelle:cmake_define WITH_SASL_SCRAM ON
# gazelle:cmake_define WITH_SASL_OAUTHBEARER ON
# gazelle:cmake_define WITH_SASL OFF
```

The `rdkafka_sasl_oauthbearer.c` file unconditionally includes OpenSSL headers:
```c
#include <openssl/evp.h>
```

However, the build configuration has:
- `WITH_SSL OFF` - SSL support is disabled
- `WITH_SASL_OAUTHBEARER ON` - SASL OAuth Bearer is enabled

This creates a conflict because SASL OAuth Bearer implementation requires OpenSSL, but SSL support is disabled.

## Suggested Fixes

### Option 1: Enable SSL Support
Change the configuration to enable SSL when using SASL OAuth Bearer:
```starlark
# gazelle:cmake_define WITH_SSL ON
```

### Option 2: Disable SASL OAuth Bearer
If SSL is not needed, disable SASL OAuth Bearer:
```starlark
# gazelle:cmake_define WITH_SASL_OAUTHBEARER OFF
```

### Option 3: Install OpenSSL Development Headers
As a workaround, install OpenSSL development headers on the build system:
```bash
# Ubuntu/Debian
sudo apt-get install libssl-dev

# RHEL/CentOS/Fedora
sudo yum install openssl-devel
```

### Option 4: Fix Conditional Compilation
The librdkafka source should properly handle the case where `WITH_SASL_OAUTHBEARER` is enabled but `WITH_SSL` is disabled by either:
- Adding conditional compilation guards around OpenSSL includes
- Automatically enabling SSL when SASL OAuth Bearer is enabled
- Failing with a clear error message during configuration

## Environment

- OS: Linux 6.8.0-1024-aws
- Bazel version: 8.2.1
- Repository: goniz/gazelle-foreign-cc

## Additional Notes

The original command provided had two typos:
1. `thirdpart` should be `thirdparty`
2. `librdkafk` should be `librdkafka`

## Resolution

The issue was successfully resolved by implementing **Option 2: Disable SASL OAuth Bearer**.

### Steps Taken:

1. **First attempt**: Disabled SASL OAuth Bearer by changing `WITH_SASL_OAUTHBEARER ON` to `WITH_SASL_OAUTHBEARER OFF` in `examples/thirdparty/librdkafka/BUILD.bazel`.

2. **Second issue discovered**: The build then failed on `rdkafka_sasl_scram.c` with a similar error because SASL SCRAM also requires OpenSSL.

3. **Final fix**: Disabled both SASL OAuth Bearer and SASL SCRAM:
   ```starlark
   # gazelle:cmake_define WITH_SASL_SCRAM OFF
   # gazelle:cmake_define WITH_SASL_OAUTHBEARER OFF
   ```

4. **Regenerated build files**: Ran `bazel run :gazelle` to update the build configuration.

5. **Successful build**: The build completed successfully, generating:
   - `librdkafka.a` - static library for C API
   - `librdkafka.so` - shared library for C API
   - `librdkafka++.a` - static library for C++ API
   - `librdkafka++.so` - shared library for C++ API

### Key Findings:

- Both SASL OAuth Bearer and SASL SCRAM authentication mechanisms in librdkafka have a hard dependency on OpenSSL.
- If SSL support is disabled (`WITH_SSL OFF`), both `WITH_SASL_OAUTHBEARER` and `WITH_SASL_SCRAM` must also be disabled.
- The gazelle-foreign-cc tool correctly regenerates the build files when these configuration directives are changed, removing the problematic source files from the compilation.