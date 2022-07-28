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

#include <mutex>
#include <string>
#include <sstream>

using LedColorsVector = std::vector<CorsairLedColor>;
using AvailableKeys = std::unordered_map<int /*device index*/, LedColorsVector>;
std::mutex m;

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

AvailableKeys getAvailableKeys(std::vector<int> colorsByGo)
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

void performPulseEffect(AvailableKeys &availableKeys)
{
	const auto timePerFrame = 25;
	//std::lock_guard<std::mutex> guard(myMutex);
	m.lock();
		for (auto &ledColorsByDeviceIndex : availableKeys) {
			auto &deviceIndex = ledColorsByDeviceIndex.first;
			auto &ledColorsVec = ledColorsByDeviceIndex.second;
			CorsairSetLedsColorsBufferByDeviceIndex(deviceIndex, static_cast<int>(ledColorsVec.size()), ledColorsVec.data());
		}
		CorsairSetLedsColorsFlushBufferAsync(nullptr, nullptr);
	m.unlock();
		std::this_thread::sleep_for(std::chrono::milliseconds(timePerFrame));
}


void consoleInputs(AvailableKeys* pt)
{
	bool checker = true;
	while (checker == true) {
		checker = false;
		std::string test;
		std::cin >> test;
		if (test.size() > 2) {
			checker = true;
			std::vector<int> vect;
			std::stringstream ss(test);
			for (int i; ss >> i;) {
				vect.push_back(i);
				if (ss.peek() == ',')
					ss.ignore();
			}
			if (vect.size() == 9) {
				//std::lock_guard<std::mutex> guard(myMutex);
				m.lock();
					*pt = getAvailableKeys(vect);
				m.unlock();
			}
			
		}
		else {
			checker = true;
		}
	}	
}

int main()
{
	
	CorsairPerformProtocolHandshake();
	if (const auto error = CorsairGetLastError()) {
		std::cout << "Handshake failed: " << toString(error) << "\nPress any key to quit." << std::endl;
		getchar();
		return -1;
	}

	std::atomic_bool continueExecution{ true };
	std::vector<int> vect(9);

	auto availableKeys = getAvailableKeys(vect);
	auto* ptr = &availableKeys;
	
	if (availableKeys.empty()) {
		return 1;
	}

	std::thread consoleInputs(consoleInputs, ptr);

	while (true) {
		performPulseEffect(availableKeys);
	}

	consoleInputs.join();

	return 0;
}
