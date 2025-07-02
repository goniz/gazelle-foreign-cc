#include <curl/curl.h>
#include <iostream>
#include <fstream>
#include <string>
#include <iomanip>

// Structure to hold download progress data
struct ProgressData {
    double lastProgress;
};

// Callback function to write data to file
size_t WriteFileCallback(void* contents, size_t size, size_t nmemb, std::ofstream* file) {
    size_t totalSize = size * nmemb;
    file->write((char*)contents, totalSize);
    return totalSize;
}

// Callback function to display download progress
int ProgressCallback(void* userData, curl_off_t dlTotal, curl_off_t dlNow, 
                     curl_off_t ulTotal, curl_off_t ulNow) {
    ProgressData* progress = (ProgressData*)userData;
    
    if (dlTotal > 0) {
        double percentage = (double)dlNow / dlTotal * 100.0;
        
        // Only update display if progress changed by at least 1%
        if (percentage - progress->lastProgress >= 1.0 || percentage == 100.0) {
            progress->lastProgress = percentage;
            
            std::cout << "\rDownload Progress: " << std::fixed << std::setprecision(1) 
                      << percentage << "% "
                      << "[" << dlNow << "/" << dlTotal << " bytes]" << std::flush;
        }
    }
    
    return 0; // Return 0 to continue, non-zero to abort
}

int main(int argc, char* argv[]) {
    std::cout << "libcurl File Download Example" << std::endl;
    std::cout << "=============================" << std::endl;

    if (argc != 3) {
        std::cerr << "Usage: " << argv[0] << " <URL> <output_filename>" << std::endl;
        std::cerr << "Example: " << argv[0] << " http://www.w3.org/WAI/ER/tests/xhtml/testfiles/resources/pdf/dummy.pdf dummy.pdf" << std::endl;
        return 1;
    }

    const char* url = argv[1];
    const char* outputFile = argv[2];

    // Initialize libcurl
    curl_global_init(CURL_GLOBAL_DEFAULT);

    // Create a curl handle
    CURL* curl = curl_easy_init();
    if (curl) {
        // Open output file
        std::ofstream file(outputFile, std::ios::binary);
        if (!file.is_open()) {
            std::cerr << "Failed to open output file: " << outputFile << std::endl;
            curl_easy_cleanup(curl);
            curl_global_cleanup();
            return 1;
        }

        ProgressData progressData = {0.0};

        std::cout << "Downloading: " << url << std::endl;
        std::cout << "Saving to: " << outputFile << std::endl << std::endl;

        // Set curl options
        curl_easy_setopt(curl, CURLOPT_URL, url);
        curl_easy_setopt(curl, CURLOPT_WRITEFUNCTION, WriteFileCallback);
        curl_easy_setopt(curl, CURLOPT_WRITEDATA, &file);
        curl_easy_setopt(curl, CURLOPT_FOLLOWLOCATION, 1L);
        curl_easy_setopt(curl, CURLOPT_TIMEOUT, 30L);
        
        // Enable progress meter
        curl_easy_setopt(curl, CURLOPT_NOPROGRESS, 0L);
        curl_easy_setopt(curl, CURLOPT_XFERINFOFUNCTION, ProgressCallback);
        curl_easy_setopt(curl, CURLOPT_XFERINFODATA, &progressData);

        // Set user agent
        curl_easy_setopt(curl, CURLOPT_USERAGENT, "libcurl-download-example/1.0");

        // Perform the download
        CURLcode res = curl_easy_perform(curl);
        std::cout << std::endl; // New line after progress

        if (res != CURLE_OK) {
            std::cerr << "Download failed: " << curl_easy_strerror(res) << std::endl;
            file.close();
            std::remove(outputFile); // Remove partial file
        } else {
            // Get download info
            double downloadSize;
            double downloadTime;
            double downloadSpeed;
            
            curl_easy_getinfo(curl, CURLINFO_SIZE_DOWNLOAD, &downloadSize);
            curl_easy_getinfo(curl, CURLINFO_TOTAL_TIME, &downloadTime);
            curl_easy_getinfo(curl, CURLINFO_SPEED_DOWNLOAD, &downloadSpeed);

            std::cout << std::endl << "Download completed successfully!" << std::endl;
            std::cout << "Downloaded: " << std::fixed << std::setprecision(2) 
                      << downloadSize / 1024.0 << " KB" << std::endl;
            std::cout << "Time taken: " << downloadTime << " seconds" << std::endl;
            std::cout << "Average speed: " << downloadSpeed / 1024.0 << " KB/s" << std::endl;
        }

        file.close();
        curl_easy_cleanup(curl);
    } else {
        std::cerr << "Failed to initialize curl" << std::endl;
        curl_global_cleanup();
        return 1;
    }

    curl_global_cleanup();
    return 0;
}