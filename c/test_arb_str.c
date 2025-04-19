#include <stdio.h>
#include <arb.h>

int main() {
    arb_t x;
    arb_init(x);

    const char *s = "1.23456864936982369264298462339e11";
    int ok = arb_set_str(x, s, 256);  // 256-bit precision

    if (!ok) {
        printf("✅ Success parsing: %s\n", s);
        arb_printd(x, 20);  // print with 20 digits
        printf("\n");
    } else {
        printf("❌ Failed to parse: %s\n", s);
    }

    arb_clear(x);
    return 0;
}

