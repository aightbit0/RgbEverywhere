/******************************************************************************
**
** File      color_pulse_by_device_index.cpp
** Author    Illia Okonskyi
** Copyright (c) 2021, Corsair Memory, Inc. All Rights Reserved.
**
** This file is part of Corsair SDK Lighting Effects.
**
******************************************************************************/

#ifdef __APPLE__
#include <CUESDK/CUESDK.h>
#else
#include <CUESDK.h>
#endif

#include <iostream>
#include <atomic>
#include <thread>
#include <vector>
#include <unordered_map>
#include <cmath>

using LedColorsVector = std::vector<CorsairLedColor>;
using AvailableKeys = std::unordered_map<int /*device index*/, LedColorsVector>;

const char* toString(CorsairError error)
{
	switch (error) {
	case CE_Success:
		return "CE_Success";
	case CE_ServerNotFound:
		return "CE_ServerNotFound";
	case CE_NoControl:
		return "CE_NoControl";
	case CE_ProtocolHandshakeMissing:
		return "CE_ProtocolHandshakeMissing";
	case CE_IncompatibleProtocol:
		return "CE_IncompatibleProtocol";
	case CE_InvalidArguments:
		return "CE_InvalidArguments";
	default:
		return "unknown error";
	}
}

AvailableKeys getAvailableKeys(int colorsByGo[])
{
	AvailableKeys availableKeys;

	for (int i = 0; i < CorsairGetDeviceCount(); i++)
	{
		const auto ledPositions = CorsairGetLedPositionsByDeviceIndex(i);
		LedColorsVector keys;
		for (auto i = 0; i < ledPositions->numberOfLed; i++) {
			const auto ledId = ledPositions->pLedPosition[i].ledId;

			if (i % 2 == 0) {
				keys.push_back(CorsairLedColor{ ledId, colorsByGo[0], colorsByGo[1], colorsByGo[2]});
			}
			else if (i % 3 == 0) {
				
				keys.push_back(CorsairLedColor{ ledId, colorsByGo[3], colorsByGo[4], colorsByGo[5] });
			}
			else {
				keys.push_back(CorsairLedColor{ ledId, colorsByGo[6], colorsByGo[7], colorsByGo[8] });
			}
			
		}
		availableKeys[i] = keys;
	}
	
	return availableKeys;
}

void performPulseEffect(int waveDuration, AvailableKeys &availableKeys)
{
	const auto timePerFrame = 25;

		for (auto &ledColorsByDeviceIndex : availableKeys) {
			auto &deviceIndex = ledColorsByDeviceIndex.first;
			auto &ledColorsVec = ledColorsByDeviceIndex.second;
			CorsairSetLedsColorsBufferByDeviceIndex(deviceIndex, static_cast<int>(ledColorsVec.size()), ledColorsVec.data());
		}
		CorsairSetLedsColorsFlushBufferAsync(nullptr, nullptr);
		std::this_thread::sleep_for(std::chrono::milliseconds(timePerFrame));
}


void consoleInputs(AvailableKeys* pt)
{
	bool checker = true;
	int color2[9] = { 0,0,0,0,0,0,0,0,0};
	int i = 0;
	while (checker == true) {
		checker = false;
		//std::this_thread::sleep_for(std::chrono::seconds(5));
		std::string test;
		std::cin >> test;
		std::cout << "das ist der wert: " << test << std::endl;
		if (test != "") {
			checker = true;
			color2[i] = atoi(test.c_str());
			if (i < 9) {
				i++;
				if (i == 9) {
					for (int w = 0; w < 9; w++)
					{
						std::cout << color2[w] << std::endl;
					}
					i = 0;
					*pt = getAvailableKeys(color2);
				}
			}
		}
		else {
			*pt = getAvailableKeys(color2);
		}
	}	
}

int main(int argc, char**argv)
{
	CorsairPerformProtocolHandshake();
	if (const auto error = CorsairGetLastError()) {
		std::cout << "Handshake failed: " << toString(error) << "\nPress any key to quit." << std::endl;
		getchar();
		return -1;
	}

	std::atomic_int waveDuration{ 2500 };
	std::atomic_bool continueExecution{ true };

	int color1[9] = {0,0,0,0,0,0,0,0,0};

	std::cout << "You have entered " << argc
	<< " arguments:" << "\n";

	if (argc != 1) {
		for (int i = 0; i < argc; ++i) {
			std::cout << argv[i] << "\n";
			if (i != 0) {
				std::string arg1(argv[i]);
				std::cout << arg1 << "\n";
				color1[i - 1] = atoi(arg1.c_str());
			}
		}
	}

	auto availableKeys = getAvailableKeys(color1);

	AvailableKeys* ptr = &availableKeys;
	
	if (availableKeys.empty()) {
		return 1;
	}

	std::thread consoleInputs(consoleInputs, ptr);

	while (true) {
		performPulseEffect(waveDuration, availableKeys);
	}

	consoleInputs.join();

	return 0;
}
