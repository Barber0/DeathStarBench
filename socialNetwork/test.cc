#include <iostream>
#include <string>
#include <cstdio>

using namespace std;

string intToString(int v)
{
    char buf[32] = {0};
    snprintf(buf, sizeof(buf), "%u", v);

    string str = buf;
    return str;
}

int main(int argc, char **argv)
{
    string mystr = "abc";
    mystr = mystr + intToString(123);
    cout << mystr << endl;
}