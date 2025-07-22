#include <curl/curl.h>
#include <iostream>
#include <string>

// Callback function to handle the response data
size_t WriteCallback(void* contents, size_t size, size_t nmemb, std::string* response) {
    size_t totalSize = size * nmemb;
    response->append((char*)contents, totalSize);
    return totalSize;
}

int main(int argc, char* argv[]) {
    std::cout << "libcurl HTTP GET Example" << std::endl;
    std::cout << "========================" << std::endl;

    // Initialize libcurl
    curl_global_init(CURL_GLOBAL_DEFAULT);

    // Create a curl handle
    CURL* curl = curl_easy_init();
    if (curl) {
        std::string response;
        
        // Set the URL to fetch
        const char* url = "http://httpbin.org/get";
        if (argc > 1) {
            url = argv[1];
        }
        
        std::cout << "Fetching: " << url << std::endl << std::endl;
        
        // Set options
        curl_easy_setopt(curl, CURLOPT_URL, url);
        curl_easy_setopt(curl, CURLOPT_WRITEFUNCTION, WriteCallback);
        curl_easy_setopt(curl, CURLOPT_WRITEDATA, &response);
        curl_easy_setopt(curl, CURLOPT_FOLLOWLOCATION, 1L);
        curl_easy_setopt(curl, CURLOPT_TIMEOUT, 10L);
        
        // Perform the request
        CURLcode res = curl_easy_perform(curl);
        
        if (res != CURLE_OK) {
            std::cerr << "curl_easy_perform() failed: " << curl_easy_strerror(res) << std::endl;
        } else {
            // Get the HTTP response code
            long http_code = 0;
            curl_easy_getinfo(curl, CURLINFO_RESPONSE_CODE, &http_code);
            
            std::cout << "HTTP Response Code: " << http_code << std::endl;
            std::cout << "Response Body:" << std::endl;
            std::cout << response << std::endl;
        }
        
        // Cleanup
        curl_easy_cleanup(curl);
    } else {
        std::cerr << "Failed to initialize curl" << std::endl;
        return 1;
    }
    
    // Global cleanup
    curl_global_cleanup();
    
    return 0;
}