/**************************************************************************/
/*!
  MCP427DAC -> ARDUINO

  VDD -> Vin
  GND -> GND
  SCL -> A5
  SDA -> A4

*/
/**************************************************************************/
#include <Wire.h>
#include <Adafruit_MCP4725.h>

Adafruit_MCP4725 dac;
String sdata = ""; // Initialised to nothing.
bool started = false;
void setup(void) {
  Serial.begin(9600);

  // For Adafruit MCP4725A1 the address is 0x62 (default) or 0x63 (ADDR pin tied to VCC)
  // For MCP4725A0 the address is 0x60 or 0x61
  // For MCP4725A2 the address is 0x64 or 0x65
  dac.begin(0x60);
  pinMode(2, OUTPUT);
  pinMode(3, OUTPUT);

  Serial.println("Begin");
}

void loop(void) {
  if (started == false) {
    started = true;
    dac.setVoltage(0, 1);
    digitalWrite(2, LOW);
    digitalWrite(3, LOW);
  }
  byte ch;
  if (Serial.available()) {
    ch = Serial.read();
    sdata += (char)ch;
    if (ch == '\n') {
      sdata.trim();
      if (sdata.indexOf("voltage") > -1) {
        sdata.remove(0, 7);
        float newVal = sdata.toFloat();
        Serial.println(newVal);
        // set voltage
        float newVoltage = round(4095.0/3.34/2.0* newVal);
        if (newVoltage > 4095) {
          newVoltage = 4095;
        }
        uint16_t newVolts = uint16_t(newVoltage);
        dac.setVoltage(newVolts, 1);
        Serial.print("volts: ");
        Serial.println(newVolts);
      } else if (sdata.indexOf("sol1") > -1) {
        if (sdata.indexOf("on") > -1) {
          digitalWrite(2, HIGH);
          Serial.println("sol1on");
        } else {
          digitalWrite(2, LOW);
          Serial.println("sol1off");
        }
      } else if (sdata.indexOf("sol2") > -1) {
        if (sdata.indexOf("on") > -1) {
          digitalWrite(3, HIGH);
          Serial.println("sol2on");
        } else {
          digitalWrite(3, LOW);
          Serial.println("sol2off");
        }
      } else {
        Serial.println("?");
      }
      sdata = "";
    }
  }
}
