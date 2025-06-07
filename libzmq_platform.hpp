#ifndef __ZMQ_PLATFORM_HPP_INCLUDED__
#define __ZMQ_PLATFORM_HPP_INCLUDED__

/* Basic platform configuration for libzmq built with Bazel */
/* This file provides minimal platform definitions to replace the CMake-generated platform.hpp */

/* Condition variable implementation */
#define ZMQ_USE_CV_IMPL_STL11

/* I/O thread polling - use epoll on Linux, kqueue on BSD/macOS, select as fallback */
#if defined(__linux__)
  #define ZMQ_IOTHREAD_POLLER_USE_EPOLL
  #define ZMQ_IOTHREAD_POLLER_USE_EPOLL_CLOEXEC
  #define ZMQ_HAVE_LINUX
#elif defined(__APPLE__)
  #define ZMQ_IOTHREAD_POLLER_USE_KQUEUE
  #define ZMQ_HAVE_OSX
#elif defined(__FreeBSD__) || defined(__FreeBSD_kernel__)
  #define ZMQ_IOTHREAD_POLLER_USE_KQUEUE
  #define ZMQ_HAVE_FREEBSD
#elif defined(__OpenBSD__)
  #define ZMQ_IOTHREAD_POLLER_USE_KQUEUE
  #define ZMQ_HAVE_OPENBSD
#elif defined(__NetBSD__)
  #define ZMQ_IOTHREAD_POLLER_USE_KQUEUE
  #define ZMQ_HAVE_NETBSD
#else
  #define ZMQ_IOTHREAD_POLLER_USE_SELECT
#endif

/* API polling implementation */
#define ZMQ_POLL_BASED_ON_POLL

/* Enable common POSIX features */
#define HAVE_FORK
#define HAVE_CLOCK_GETTIME
#define ZMQ_HAVE_UIO

/* Enable common socket features */
#define ZMQ_HAVE_EVENTFD
#define ZMQ_HAVE_EVENTFD_CLOEXEC
#define ZMQ_HAVE_O_CLOEXEC
#define ZMQ_HAVE_SOCK_CLOEXEC
#define ZMQ_HAVE_SO_KEEPALIVE
#define ZMQ_HAVE_TCP_KEEPCNT
#define ZMQ_HAVE_TCP_KEEPIDLE
#define ZMQ_HAVE_TCP_KEEPINTVL
#define ZMQ_HAVE_TCP_KEEPALIVE

/* Pthread features */
#define ZMQ_HAVE_PTHREAD_SETNAME_2
#define ZMQ_HAVE_PTHREAD_SET_AFFINITY

/* String functions */
#define HAVE_STRNLEN
#define ZMQ_HAVE_STRLCPY

/* Enable IPC transport */
#define ZMQ_HAVE_IPC

/* Use built-in SHA1 implementation */
#define ZMQ_USE_BUILTIN_SHA1

/* WebSocket support */
#define ZMQ_HAVE_WS

/* Platform-specific includes based on compiler macros */
#ifdef _AIX
  #define ZMQ_HAVE_AIX
#endif

#if defined __ANDROID__
  #define ZMQ_HAVE_ANDROID
#endif

#if defined __CYGWIN__
  #define ZMQ_HAVE_CYGWIN
#endif

#if defined __MINGW32__
  #define ZMQ_HAVE_MINGW32
#endif

#if defined(__DragonFly__)
  #define ZMQ_HAVE_FREEBSD
  #define ZMQ_HAVE_DRAGONFLY
#endif

#if defined __hpux
  #define ZMQ_HAVE_HPUX
#endif

#if defined __QNXNTO__
  #define ZMQ_HAVE_QNXNTO
#endif

#if defined(sun) || defined(__sun)
  #define ZMQ_HAVE_SOLARIS
#endif

#if defined __VMS
  #define ZMQ_HAVE_OPENVMS
  #undef ZMQ_HAVE_IPC
#endif

/* Set cache line size to a reasonable default */
#define ZMQ_CACHELINE_SIZE 64

#endif