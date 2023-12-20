#include <stdio.h>

#define printValue(x) _Generic((x), \
    int: printInt, \
    double: printDouble, \
    char*: printString \
)(x)

void printInt(int x
,int y) {
    printf("Integer: %d\n", x);
}

void printDouble(double x) {
    printf("Double: %lf\n", x);
}

void printString(char* x) {
    printf("String: %s\n", x);
}

int main() {
    printValue(42);
    printValue(3.14);
    printValue("Hello");

    return 0;
}

