#ifndef __ZMQ_H_INCLUDED__
#define __ZMQ_H_INCLUDED__

#ifdef __cplusplus
extern "C" {
#endif

/*  Version macros for compile-time API version detection                     */
#define ZMQ_VERSION_MAJOR 4
#define ZMQ_VERSION_MINOR 3
#define ZMQ_VERSION_PATCH 4

/*  Context options                                                           */
#define ZMQ_IO_THREADS 1
#define ZMQ_MAX_SOCKETS 2

/*  Socket types                                                              */
#define ZMQ_PAIR 0
#define ZMQ_PUB 1
#define ZMQ_SUB 2
#define ZMQ_REQ 3
#define ZMQ_REP 4
#define ZMQ_DEALER 5
#define ZMQ_ROUTER 6
#define ZMQ_PULL 7
#define ZMQ_PUSH 8

/*  Socket options                                                            */
#define ZMQ_AFFINITY 4
#define ZMQ_IDENTITY 5
#define ZMQ_SUBSCRIBE 6
#define ZMQ_UNSUBSCRIBE 7

/*  Send/recv options                                                         */
#define ZMQ_DONTWAIT 1
#define ZMQ_SNDMORE 2

typedef struct zmq_msg_t {unsigned char _ [64];} zmq_msg_t;

/*  0MQ context API                                                           */
void *zmq_ctx_new (void);
int zmq_ctx_term (void *context);
int zmq_ctx_shutdown (void *context);
int zmq_ctx_set (void *context, int option, int optval);
int zmq_ctx_get (void *context, int option);

/*  0MQ socket API                                                            */
void *zmq_socket (void *context, int type);
int zmq_close (void *socket);
int zmq_setsockopt (void *socket, int option_name, const void *option_value, size_t option_len);
int zmq_getsockopt (void *socket, int option_name, void *option_value, size_t *option_len);
int zmq_bind (void *socket, const char *endpoint);
int zmq_connect (void *socket, const char *endpoint);
int zmq_unbind (void *socket, const char *endpoint);
int zmq_disconnect (void *socket, const char *endpoint);

/*  0MQ message API                                                           */
int zmq_msg_init (zmq_msg_t *msg);
int zmq_msg_init_size (zmq_msg_t *msg, size_t size);
int zmq_msg_init_data (zmq_msg_t *msg, void *data, size_t size, void (*ffn) (void *data, void *hint), void *hint);
size_t zmq_msg_size (const zmq_msg_t *msg);
void *zmq_msg_data (zmq_msg_t *msg);
int zmq_msg_close (zmq_msg_t *msg);

/*  0MQ send/receive API                                                      */
int zmq_msg_send (zmq_msg_t *msg, void *socket, int flags);
int zmq_msg_recv (zmq_msg_t *msg, void *socket, int flags);
int zmq_send (void *socket, const void *buf, size_t len, int flags);
int zmq_recv (void *socket, void *buf, size_t len, int flags);

/*  0MQ polling API                                                           */
typedef struct
{
    void *socket;
    int fd;
    short events;
    short revents;
} zmq_pollitem_t;

int zmq_poll (zmq_pollitem_t *items, int nitems, long timeout);

/*  Built-in message proxy                                                    */
int zmq_proxy (void *frontend, void *backend, void *capture);

/*  Utility functions                                                         */
void zmq_version (int *major, int *minor, int *patch);
int zmq_errno (void);
const char *zmq_strerror (int errnum);

#ifdef __cplusplus
}
#endif

#endif