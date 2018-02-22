#ifndef __C_GO_SYMC_H__
#define __C_GO_SYMC_H__

#include <stddef.h>
#include <stdint.h>
#include <stdbool.h>

#ifdef __cplusplus
extern "C" {
#endif

typedef void* symc_t;
typedef void* symc_result_t;

extern    symc_t symc_create(const char* vbucket_conf, const char* vbucket_name);

extern    void symc_destory(symc_t t);

extern    bool symc_get(symc_t t, const char ** keys, int keys_len, char** values, int* values_len);
extern    symc_result_t symc_get_result(symc_t t, const char ** keys, int keys_len);
extern    bool symc_set(symc_t t, const char ** keys, int keys_len, char** values, int values_len);


extern  void  symc_result_start(symc_result_t t);
extern  const char * symc_result_current_key(symc_result_t t);
extern  const char * symc_result_current_val(symc_result_t t);
extern  void  symc_result_current_next(symc_result_t t);
extern  bool  symc_result_is_end(symc_result_t t);
extern  bool  symc_result_destory(symc_result_t t);

#ifdef __cplusplus
}
#endif
#endif // __C_GO_SYMC_H__
