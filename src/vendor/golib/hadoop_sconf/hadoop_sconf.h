#ifndef __HADOOP_SCONF_H_1_
#define __HADOOP_SCONF_H_1_

#ifdef __cplusplus
extern "C" {
#endif

typedef void*  hadoop_sconf_result_t ;
hadoop_sconf_result_t raw_do_hadoop_sconf_post(const char* server_url);
const char* get_data(hadoop_sconf_result_t result);
int get_data_len(hadoop_sconf_result_t result);
void destory_result(hadoop_sconf_result_t result);

#ifdef __cplusplus
}
#endif

#endif // _HADOOP_SCONF_H_
