#include "boolexp_cgo.h"
#include "msg.pb.h"
#include "boolexp/boolexp.h"

boolexp_t boolexp_create(const char* input) {
     boolexp::BoolExp * be = new boolexp::BoolExp;
     if (!be->Init(input, false)) {
        delete be;
        return (boolexp_t)NULL;           
     }
     return (boolexp_t)be;
}

void boolexp_destory(boolexp_t be) {
    if (!be) return;
    boolexp::BoolExp* ptr = (boolexp::BoolExp*)be;
    delete ptr;
}

bool boolexp_process(boolexp_t  be, const void * data, int data_len, void** out, int* out_len) {
   if (be == NULL|| data == NULL || out == NULL || out_len == NULL) {
        return false;
   }
    boolexp::Request req;
    boolexp::Response resp;
    if (!req.ParseFromArray(data, data_len)) {
        *out = NULL;
        *out_len = 0;
        std::cerr << "ParseFromArray failed,in_len=" << data_len << std::endl;
        return false;
    }
    
    typedef ::google::protobuf::Map< ::std::string, ::std::string > gmap;

    for (int i = 0; i< req.asks_size(); ++i) {
        const gmap &  map = req.asks(i).conditions();
        boolexp::BoolExp * be_ptr =(boolexp::BoolExp*)be;
        boolexp::EvalResult br = be_ptr->Eval(map);
        boolexp::Ans * ans = resp.add_anss();
        ans->set_result(br.BoolResult());
    }
    
     std::string result;
     resp.SerializeToString(&result);

     *out = NULL;
     *out_len = result.size();
     if (!result.empty()) {
         *out =  (char*) malloc(result.size() + 1);
         memcpy(*out, result.data(), result.size() +1);
     }
    return true;
}
