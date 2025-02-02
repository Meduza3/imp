/*
 * Kod interpretera maszyny rejestrowej do projektu z JFTT2024
 *
 * Autor: Maciek Gębala
 * http://ki.pwr.edu.pl/gebala/
 * 2024-11-11
 * (wersja long long)
 */
#include <iostream>
#include <chrono>
#include <thread>
#include <locale>
#include <utility>
#include <vector>
#include <map>
#include <cstdlib> // rand()
#include <ctime>

#include "instructions.hh"
#include "colors.hh"

using namespace std;

// Helper: convert instruction code to string for debugging.
string ins_name(int code)
{
  switch (code)
  {
  case GET:
    return "GET";
  case PUT:
    return "PUT";
  case LOAD:
    return "LOAD";
  case STORE:
    return "STORE";
  case LOADI:
    return "LOADI";
  case STOREI:
    return "STOREI";
  case ADD:
    return "ADD";
  case SUB:
    return "SUB";
  case ADDI:
    return "ADDI";
  case SUBI:
    return "SUBI";
  case SET:
    return "SET";
  case HALF:
    return "HALF";
  case JUMP:
    return "JUMP";
  case JPOS:
    return "JPOS";
  case JZERO:
    return "JZERO";
  case JNEG:
    return "JNEG";
  case RTRN:
    return "RTRN";
  case HALT:
    return "HALT";
  default:
    return "UNKNOWN";
  }
}
void run_machine(vector<pair<int, long long>> &program)
{
  map<long long, long long> p; // pamięć maszyny

  int lr;          // licznik rozkazów (program counter)
  long long t, io; // liczniki kosztu i operacji I/O

  cout << cBlue << "Uruchamianie programu." << cReset << endl;
  lr = 0;
  t = 0;
  io = 0;
  while (true)
  {
    if (lr < 0 || lr >= (int)program.size())
    {
      cerr << cRed << "Błąd: Wywołanie nieistniejącej instrukcji nr " << lr << "." << cReset << endl;
      exit(-1);
    }

    int current_pc = lr;
    int current_instr = program[current_pc].first;
    long long current_op = program[current_pc].second;

    if (current_instr == HALT)
      break;

    // Sprawdzenie ujemnych adresów
    if (current_instr != SET &&
        current_instr != JUMP &&
        current_instr != JPOS &&
        current_instr != JZERO &&
        current_instr != JNEG &&
        current_op < 0)
    {
      cerr << cRed << "Błąd: ujemny adres pamięci." << cReset << endl;
      exit(-1);
    }

    std::this_thread::sleep_for(std::chrono::milliseconds(100));
    switch (current_instr)
    {
    case GET:
      cout << "? ";
      cin >> p[current_op];
      io += 100;
      t += 100;
      lr++;
      break;
    case PUT:
      cout << "> " << p[current_op] << endl;
      io += 100;
      t += 100;
      lr++;
      break;

    case LOAD:
      p[0] = p[current_op];
      t += 10;
      lr++;
      break;
    case STORE:
      p[current_op] = p[0];
      t += 10;
      lr++;
      break;
    case LOADI:
      p[0] = p[p[current_op]];
      t += 20;
      lr++;
      break;
    case STOREI:
      p[p[current_op]] = p[0];
      t += 20;
      lr++;
      break;

    case ADD:
      p[0] += p[current_op];
      t += 10;
      lr++;
      break;
    case SUB:
      p[0] -= p[current_op];
      t += 10;
      lr++;
      break;
    case ADDI:
      p[0] += p[p[current_op]];
      t += 20;
      lr++;
      break;
    case SUBI:
      p[0] -= p[p[current_op]];
      t += 12;
      lr++;
      break;

    case SET:
      p[0] = current_op;
      t += 50;
      lr++;
      break;
    case HALF:
      p[0] >>= 1;
      t += 5;
      lr++;
      break;

    case JUMP:
      lr += current_op;
      t += 1;
      break;
    case JPOS:
      if (p[0] > 0)
        lr += current_op;
      else
        lr++;
      t += 1;
      break;
    case JZERO:
      if (p[0] == 0)
        lr += current_op;
      else
        lr++;
      t += 1;
      break;
    case JNEG:
      if (p[0] < 0)
        lr += current_op;
      else
        lr++;
      t += 1;
      break;

    case RTRN:
      lr = p[current_op];
      t += 10;
      break;
    default:
      cerr << cRed << "Błąd: Nieznana instrukcja." << cReset << endl;
      exit(-1);
    }

    // Wydruk stanu pamięci po wykonaniu instrukcji
    cout << cYellow << "[DEBUG] Po wykonaniu PC " << current_pc << ": "
         << ins_name(current_instr) << " " << current_op << cReset << endl;
    cout << "Zawartość pamięci:" << endl;
    for (const auto &[addr, value] : p)
      cout << "  " << addr << ": " << value << endl;
  }
  cout << cBlue << "Skończono program (koszt: " << cRed << t
       << cBlue << "; w tym i/o: " << io << ")." << cReset << endl;
}