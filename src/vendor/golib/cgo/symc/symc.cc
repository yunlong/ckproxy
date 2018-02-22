#include "symc.h"
#include <osl/include/vbucket_conf.h>
#include <symcwrap/include/symc.h>
#include <symc/memcached.h>

struct symc_t_imp
{
    qh::VBucketConfManager* vbucket_conf_manager_;
    symc::Memcached* mc_;
    qh::VBucketConf vbucket_conf_;  
};


typedef std::map<std::string, std::pair<std::string, symc::Status> > ResultMap;

struct symc_result {
    std::map<std::string, std::pair<std::string, symc::Status> > result_map;
    ResultMap::iterator current;
};

bool initialize(struct symc_t_imp * imp, const qh::VBucketConf* vbucket_conf) {
    imp->vbucket_conf_ = *vbucket_conf; 
    if (!imp->vbucket_conf_.initialized()) {
        if (!imp->vbucket_conf_.Initialize()) {
            fprintf(stderr, "vbucket conf initialize error\n");
            return false;
        }
    }
    return true;
}

symc::Memcached* create_symc(qh::VBucketConf & vbucket_conf) {
    symc::Memcached* mc = NULL;
    symc::TimeoutOption timeout;
    timeout.recv    = vbucket_conf.recv_timeout_ms();
    timeout.send    = vbucket_conf.send_timeout_ms();
    timeout.connect = vbucket_conf.conn_timeout_ms() << 2;
    if (vbucket_conf.IsClusterURL() || vbucket_conf.IsClusterLocalFile()) {
        //vbucket cluster
        symc::ClusterOption op;
        if (vbucket_conf.token_key_filter().size() > 0) {
            op.key_filter     = symc::TokenKeyFilter(vbucket_conf.token_key_filter());
        }
        op.failover_delay = vbucket_conf.failover_delay();
        op.use_balance    = vbucket_conf.use_balance();
        mc = NewClusterSyncMemcached(vbucket_conf.server_address(), op, timeout);
    } else if (vbucket_conf.IsHostPort()) {
        //listening port
        std::string host;
        std::string port;
        osl::StringUtil::split(vbucket_conf.server_address(), host, port, ":");
        assert(!host.empty() && !port.empty());
        mc = NewSyncMemcached(host, atoi(port.c_str()), timeout);
    } else {
        //domain socket
        mc = NewSyncMemcached(vbucket_conf.server_address(), timeout);
    }

    if (!mc) {
        fprintf(stderr, "%s:%d create [%s] failed. vbucket_url=[%s] server_address_type=[%d]\n",
                    __FILE__, __LINE__,
                    vbucket_conf.server_address_type() == qh::VBucketConf::kClusterURL ? "NewClusterSyncMemcached" : "NewSyncMemcached",
                    vbucket_conf.server_address().c_str(), (int)vbucket_conf.server_address_type());
        return NULL;
    }

    for (int i = 0; i < 3; ++i) {
        if (mc->Init()) {
            return mc;
        }
        usleep(1000*10);
    }

    fprintf(stderr, "%s:%d create [%s] and then Initialize failed. vbucket_url=[%s] server_address_type=[%d]\n",
                __FILE__, __LINE__,
                vbucket_conf.server_address_type() == qh::VBucketConf::kClusterURL ? "NewClusterSyncMemcached" : "NewSyncMemcached",
                vbucket_conf.server_address().c_str(), (int)vbucket_conf.server_address_type());
    delete mc;
    return NULL;
}

symc_t symc_create(const char* vbucket_conf, const char* vbucket_name) {
    symc_t_imp* imp = new symc_t_imp;
    symc_t t = (symc_t)imp;
    
    imp->vbucket_conf_manager_ = new qh::VBucketConfManager;
    imp->mc_ = NULL;

    if (!imp->vbucket_conf_manager_->Initialize(vbucket_conf) ) {
        symc_destory(t);
        t=NULL;
        std::cerr << __func__ << " Initialize vbucket_conf_manager failed\n";
        return t;
    }
    
    const qh::VBucketConf* vbucket = imp->vbucket_conf_manager_->Find(vbucket_name);
    if (vbucket == NULL || !initialize(imp, vbucket)) {
        symc_destory(t);
        t=NULL;
        std::cerr << __func__ << " failed!\n" ;
        return false;
    }

    imp->mc_ = create_symc(imp->vbucket_conf_);
    if (imp->mc_ == NULL) {
        symc_destory(t);
        t=NULL;
        std::cerr << __func__ << " Initialize vbucket_conf_manager failed\n";
        return t;
    }
    return t;
}

void symc_destory(symc_t t) {
      if (t==NULL) {
        return ;
    }
    symc_t_imp * imp = (symc_t_imp*)t;
    delete imp->vbucket_conf_manager_;
    delete imp->mc_;
    delete imp;
}

void  symc_result_start(symc_result_t t) {
    symc_result *r = (symc_result*)t;
    r->current = r->result_map.begin();
}

const char * symc_result_current_key(symc_result_t t) {
    symc_result *r = (symc_result*)t;
    if (r->current == r->result_map.end()) {
        return NULL;
    }
    return r->current->first.c_str();
}

const char * symc_result_current_val(symc_result_t t) {
    symc_result *r = (symc_result*)t;
    if (r->current == r->result_map.end()) {
        return NULL;
    }
    
    ResultMap::iterator itr = r->current;
    if (itr->second.second.IsOk()) {
        return itr->second.first.c_str();
    } else if (itr->second.second.IsNotFound()) {

    } else {
        std::string errmsg;
        errmsg.append(itr->first);
        errmsg.append(":[", 2);
        errmsg.append(itr->second.second.ErrorMessage());
        errmsg.append("] ", 2);
        std::cerr << __func__ << " error: " << errmsg << std::endl;
    }

    return NULL;
}

void  symc_result_current_next(symc_result_t t) {
    symc_result *r = (symc_result*)t;
    ++r->current;
}
bool  symc_result_is_end(symc_result_t t) {
    symc_result *r = (symc_result*)t;
    return r->current == r->result_map.end();    
}

bool  symc_result_destory(symc_result_t t) {
    symc_result *r = (symc_result*)t;
    r->result_map.clear();
    delete r;
    return true;
}

symc_result_t symc_get_result(symc_t t, const char ** keys, int keys_len) {
    if (t==NULL) {
        return false;
    }
    symc_t_imp * imp = (symc_t_imp*)t;
    std::set<std::string> key_set;
    symc_result *result = new symc_result;
    std::map<std::string,int> index_map;
    
    for (int i =0; i<keys_len; i++) {
        index_map[keys[i]] = i;
        key_set.insert(keys[i]);
    }
    imp->mc_->MultiGet(key_set, &(result->result_map));
    return (symc_result_t)result;
}

bool symc_get(symc_t t, const char ** keys, int keys_len, char** values, int* values_len) {
    if (t==NULL || values==NULL || values_len == NULL) {
        return false;
    }
    symc_t_imp * imp = (symc_t_imp*)t;
    std::set<std::string> key_set;
    typedef std::map<std::string, std::pair<std::string, symc::Status> > ResultMap;
    ResultMap result_map;
    std::map<std::string,int> index_map;
    
    for (int i =0; i<keys_len; i++) {
        index_map[keys[i]] = i;
        key_set.insert(keys[i]);
    }

    imp->mc_->MultiGet(key_set, &result_map);
    //
    std::string errmsg;
    for(ResultMap::iterator itr = result_map.begin(); itr!=result_map.end(); ++itr) {
        std::string val;
        if (itr->second.second.IsOk()) {
            val = itr->second.first;
        } else if (itr->second.second.IsNotFound()) {
            
        } else {
            errmsg.append(itr->first);
            errmsg.append(":[", 2);
            errmsg.append(itr->second.second.ErrorMessage());
            errmsg.append("] ", 2);
            std::cerr << __func__ << " error: " << errmsg << std::endl;
        }
        int id = index_map[itr->first];
        values[id] = (char*)malloc(val.length() + 1);
        strcpy(values[id], val.c_str());
    }
    return true;
}

bool symc_set(symc_t t, const char ** keys, int keys_len, char** values, int values_len) {
    if (t==NULL) {
        return false;
    }
    symc_t_imp * imp = (symc_t_imp*)t;
    std::string errmsg;
    int hastry = 0;
    int retrytimes = 2;
    for (int i =0; i< keys_len; i++) {
        bool ok = false;
        do {
            symc::Status s = imp->mc_->Put(keys[i], values[i]);
            if (s.IsOk()) {
                ok = true;
                break;    
            } else {
                errmsg = s.ErrorMessage();
            }
        } while (hastry++ < retrytimes);
        if (!ok) {
            std::cerr << __func__ << " set failed:" << errmsg << std::endl;
            return false;
        }
    }
    return true;
}




