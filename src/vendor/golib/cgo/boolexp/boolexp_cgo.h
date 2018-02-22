#ifndef __BOOL_EXP_CG0_H__
#define __BOOL_EXP_CG0_H__

#include <stddef.h>
#include <stdint.h>
#include <stdbool.h>

#ifdef __cplusplus
extern "C" {
#endif

typedef void* boolexp_t;
extern boolexp_t boolexp_create(const char* input);
extern void boolexp_destory(boolexp_t be);
extern bool boolexp_process(boolexp_t  be, const void * data, int data_len, void** out, int* out_len);


#ifdef __cplusplus
}
#endif

#endif //__BOOL_EXP_CG0_H__
