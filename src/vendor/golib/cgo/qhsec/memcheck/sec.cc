#include "../qhsec.h"

#include <string.h>
#include <stdio.h>
#include <stdlib.h>
#include <assert.h>

int main() {
    const char* symmetric_file = "cipher/symmetric_keys.bin";
    const char* asymmetric_file_client = "cipher/client_asymmetric_keys.ini";
    const char* asymmetric_file_server = "cipher/server_asymmetric_keys.ini";

    qhsec_handle_t npp_config_cli = qhsec_create_handler(symmetric_file, "client", asymmetric_file_client);
    qhsec_handle_t npp_config_srv = qhsec_create_handler(symmetric_file, "server", asymmetric_file_server);

    qhsec_c_packer_t packer_cli = qhsec_create_c_packer(npp_config_cli, 6);
    qhsec_c_packer_set_option(packer_cli, "asymmetric_method", 4);
    qhsec_c_packer_set_option(packer_cli, "symmetric_method", 2);
    qhsec_c_packer_set_option(packer_cli, "symmetric_key_no", 8621);
    qhsec_c_packer_set_option(packer_cli, "net_method", 13);

    qhsec_s_unpacker_t unpacker_srv = qhsec_create_s_unpacker(npp_config_srv);

    const char* data_client = "0123456789";
    int len_client = 10;
    void* packed_data_client = qhsec_c_pack(packer_cli, data_client, &len_client);
    void* unpacked_data_server = qhsec_s_unpack(unpacker_srv, packed_data_client, &len_client);
    assert(memcmp(data_client, unpacked_data_server, len_client) == 0);
    
    //printf("%s - %s - %d\n", data_client, (char*)unpacked_data_server, len_client);

	const char* data_server = "9876543210";
    int len_server = 10;
    void* packed_data_server = qhsec_s_pack(unpacker_srv, data_server, &len_server);
    void* unpacked_data_client = qhsec_c_unpack(packer_cli, packed_data_server, &len_server);
    assert(memcmp(data_server, unpacked_data_client, len_server) == 0);

    //printf("%s - %s - %d\n", data_server, (char*)unpacked_data_client, len_server);

    assert(unpacked_data_client != NULL);
    assert(packed_data_server != NULL);

    free(unpacked_data_client);
    free(packed_data_server);

    qhsec_destroy_c_packer(packer_cli);
    qhsec_destroy_s_unpacker(unpacker_srv);

    qhsec_destroy_handler(npp_config_srv);
    qhsec_destroy_handler(npp_config_cli);

    return 0;
}
