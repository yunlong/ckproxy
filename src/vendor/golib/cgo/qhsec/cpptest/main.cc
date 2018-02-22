#include <string.h>
#include <stdlib.h>
#include <stdio.h>

#include <netproto/include/qhsec.h>

void v6()
{
    qhsec_handle_t chandler = qhsec_create_handler("symmetric_keys.bin", "", "client_asymmetric_keys.ini");//client
    qhsec_handle_t shandler = qhsec_create_handler("symmetric_keys.bin", "", "server_asymmetric_keys.ini");//server

    const char* data = "HELLO 360";
    int len = strlen(data);

    //client pack using V6
    qhsec_c_packer_t cp = qhsec_create_c_packer(chandler, 6);//v6
    qhsec_c_packer_set_option(cp, "asymmetric_method", 4); // NaclEC
    qhsec_c_packer_set_option(cp, "symmetric_method", 2); // IDEA

    int slen = len;
    void* cipher = qhsec_c_pack(cp, data, &slen); // DONOT need be free(ed)

    //server unpack
    qhsec_s_unpacker_t sp = qhsec_create_s_unpacker(shandler);
    int plen = slen;
    void* plain = qhsec_s_unpack(sp, cipher, &plen); // DONOT need be free(ed)
    if (strncmp(data, (char*)plain, len) == 0 && plen == len) {
        printf("v6 server unpack OK: %s\n", plain);
    } else {
        printf("v6 server unpack FAILED\n");
        exit(-1);
    }

    // server pack
    const char* ok = "I HATE GNU";
    len = strlen(ok);
    slen = len;
    cipher = qhsec_s_pack(sp, ok, &slen);

    // client unpack
    plen = slen;
    plain = qhsec_c_unpack(cp, cipher, &plen);
    if (strncmp(ok, (char*)plain, len) == 0 && plen == len) {
        printf("v6 client unpack OK :%s\n", plain);
    } else {
        printf("v6 client unpack FAILED\n");
        exit(-1);
    }

    // free resource
    free(cipher);
    free(plain);
    qhsec_destroy_handler(chandler);
    qhsec_destroy_handler(shandler);
    qhsec_destroy_c_packer(cp);
    qhsec_destroy_s_unpacker(sp);
}



void v11()
{
    qhsec_handle_t chandler = qhsec_create_handler("symmetric_keys.bin", "", "client_asymmetric_keys.ini");//client
    qhsec_handle_t shandler = qhsec_create_handler("symmetric_keys.bin", "", "server_asymmetric_keys.ini");//server

    const char* data = "123";
    int len = strlen(data);

    //client pack using V11
    qhsec_c_packer_t cp = qhsec_create_c_packer(chandler, 11);//v11
    qhsec_c_packer_set_option(cp, "symmetric_method", 2); // IDEA
    qhsec_c_packer_set_option(cp, "symmetric_key_no", 8621); // IDEA key no

    int slen = len;
    void* cipher = qhsec_c_pack(cp, data, &slen); // DONOT need be free(ed)


    //server unpack
    qhsec_s_unpacker_t sp = qhsec_create_s_unpacker(shandler);
    int plen = slen;
    void* plain = qhsec_s_unpack(sp, cipher, &plen); // DONOT need be free(ed)
    if (strncmp(data, (char*)plain, len) == 0 && plen == len) {
        printf("v11 server unpack OK\n");
    } else {
        printf("v11 server unpack FAILED\n");
        exit(-1);
    }

    // server pack
    const char* ok = "OK";
    len = strlen(ok);
    slen = len;
    cipher = qhsec_s_pack(sp, ok, &slen);

    // client unpack
    plen = slen;
    plain = qhsec_c_unpack(cp, cipher, &plen);
    if (strncmp(ok, (char*)plain, len) == 0 && plen == len) {
        printf("v11 client unpack OK\n");
    } else {
        printf("v11 client unpack FAILED\n");
        exit(-1);
    }

    // free resource
    free(cipher);
    free(plain);
    qhsec_destroy_handler(chandler);
    qhsec_destroy_handler(shandler);
    qhsec_destroy_c_packer(cp);
    qhsec_destroy_s_unpacker(sp);
}



void v6v11()
{
    qhsec_handle_t chandler = qhsec_create_handler("", "", "client_asymmetric_keys.ini");//client
    qhsec_handle_t shandler = qhsec_create_handler("symmetric_keys.bin", "", "server_asymmetric_keys.ini");//server

    qhsec_session_key_t key;

    const char* data = "123";
    int len = strlen(data);

    {
        //===============================v6====================================
        //client pack using V6
        qhsec_c_packer_t cp = qhsec_create_c_packer(chandler, 6);//v6
        qhsec_c_packer_set_option(cp, "asymmetric_method", 4); // NaclEC
        qhsec_c_packer_set_option(cp, "symmetric_method", 2); // IDEA

        int slen = len;
        void* cipher = qhsec_c_pack(cp, data, &slen); // DONOT need be free(ed)


        //server unpack
        qhsec_s_unpacker_t sp = qhsec_create_s_unpacker(shandler);
        int plen = slen;
        void* plain = qhsec_s_unpack(sp, cipher, &plen); // DONOT need be free(ed)
        if (strncmp(data, (char*)plain, len) == 0 && plen == len) {
            printf("%s v6 server unpack OK\n", __func__);
        } else {
            printf("%s v6 server unpack FAILED\n", __func__);
            exit(-1);
        }

        // server pack
        const char* ok = "OK";
        len = strlen(ok);
        slen = len;
        cipher = qhsec_s_pack(sp, ok, &slen);

        // client unpack
        plen = slen;
        plain = qhsec_c_unpack(cp, cipher, &plen);
        if (strncmp(ok, (char*)plain, len) == 0 && plen == len) {
            printf("%s v6 client unpack OK\n", __func__);
        } else {
            printf("%s v6 client unpack FAILED\n", __func__);
            exit(-1);
        }

        key = qhsec_get_session_key(chandler);

        // free resource
        free(cipher);
        free(plain);
        qhsec_destroy_c_packer(cp);
        qhsec_destroy_s_unpacker(sp);
    }

    //===============================v11====================================
    {
        qhsec_c_packer_t cp = qhsec_create_c_packer(chandler, 11);//v6
        qhsec_add_symmetric_key(chandler, key.symm_key_type, key.symm_key_id, key.symm_key, key.symm_key_len);
        qhsec_c_packer_set_option(cp, "symmetric_method", key.symm_key_type);
        qhsec_c_packer_set_option(cp, "symmetric_key_no", key.symm_key_id);

        int slen = len;
        void* cipher = qhsec_c_pack(cp, data, &slen); // DONOT need be free(ed)


        //server unpack
        qhsec_s_unpacker_t sp = qhsec_create_s_unpacker(shandler);
        int plen = slen;
        void* plain = qhsec_s_unpack(sp, cipher, &plen); // DONOT need be free(ed)
        if (strncmp(data, (char*)plain, len) == 0 && plen == len) {
            printf("%s v11 server unpack OK\n", __func__);
        } else {
            printf("%s v11 server unpack FAILED\n", __func__);
            exit(-1);
        }

        // server pack
        const char* ok = "OK";
        len = strlen(ok);
        slen = len;
        cipher = qhsec_s_pack(sp, ok, &slen);

        // client unpack
        plen = slen;
        plain = qhsec_c_unpack(cp, cipher, &plen);
        if (strncmp(ok, (char*)plain, len) == 0 && plen == len) {
            printf("%s v11 client unpack OK\n", __func__);
        } else {
            printf("%s v11 client unpack FAILED\n", __func__);
            exit(-1);
        }

        // free resource
        free(cipher);
        free(plain);
        qhsec_destroy_c_packer(cp);
        qhsec_destroy_s_unpacker(sp);
    }

    qhsec_destroy_handler(chandler);
    qhsec_destroy_handler(shandler);
}


int main()
{
    v6();
    v11();
    v6v11();
    return 0;
}
