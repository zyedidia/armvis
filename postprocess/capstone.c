#include <fcntl.h>
#include <capstone/capstone.h>
#include <string.h>
#include <unistd.h>
#include <stdint.h>
#include <sys/mman.h>

int main(int argc, char* argv[]) {
    csh handle;
    if (cs_open(CS_ARCH_AARCH64, CS_MODE_ARM, &handle) != CS_ERR_OK) {
        fprintf(stderr, "could not open arm64 disassembler\n");
        return 1;
    }
    size_t n_verify = 0x0ffffffffULL + 1;
    int fd = open(argv[1], O_RDWR);
    if (fd < 0) {
        perror("open");
    }

    int16_t* data = mmap(NULL, n_verify * sizeof(int16_t), PROT_READ | PROT_WRITE, MAP_SHARED, fd, 0);
    if (data == (int16_t*) -1) {
        perror("mmap");
    }

    size_t size = sizeof(uint32_t);
    for (size_t i = 0; i < n_verify; i++) {
        if (i % 10000000 == 0) {
            printf("%.3f\n", (float) i / (float) n_verify * 100);
        }
        if (data[i] != -1) {
            cs_insn* insn;
            uint32_t idat = (uint32_t) i;
            size_t count = cs_disasm(handle, (const uint8_t*) &idat, size, 0, 1, &insn);
            if (count != 1) {
                data[i] = -1;
                continue;
            }
            cs_free(insn, 1);
        }
    }
    close(fd);

    return 0;
}
