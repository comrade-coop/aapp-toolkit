#include <iostream>
#include <sstream>
#include <fstream>
#include <string>
#include <cstring>
#include <cstdlib>
#include <cstdio>
#include <unistd.h>
#include <sys/socket.h>
#include <netinet/in.h>
#include <algorithm>

// Include nlohmann/json for JSON processing.
#include <nlohmann/json.hpp>
using json = nlohmann::json;

using namespace std;

const int PORT = 12345;

// Reads from the client socket until the end-of-headers marker ("\r\n\r\n") is found.
string readHeaders(int client_socket) {
    const size_t BUF_SIZE = 4096;
    char buffer[BUF_SIZE];
    string headers;
    ssize_t bytes_read;
    while ((bytes_read = read(client_socket, buffer, BUF_SIZE)) > 0) {
        headers.append(buffer, bytes_read);
        if (headers.find("\r\n\r\n") != string::npos)
            break;
    }
    return headers;
}

// Extracts the Content-Length header from the headers string.
// Returns -1 if not found.
int getContentLength(const string &headers) {
    size_t pos = headers.find("Content-Length:");
    if (pos == string::npos)
        return -1;
    pos += strlen("Content-Length:");
    size_t endPos = headers.find("\r\n", pos);
    if (endPos == string::npos)
        return -1;
    string lengthStr = headers.substr(pos, endPos - pos);
    // Remove whitespace.
    lengthStr.erase(remove_if(lengthStr.begin(), lengthStr.end(), ::isspace), lengthStr.end());
    return stoi(lengthStr);
}

// Reads exactly 'length' bytes from the socket.
string readBody(int client_socket, int length) {
    string body;
    const size_t BUF_SIZE = 4096;
    char buffer[BUF_SIZE];
    int remaining = length;
    while (remaining > 0) {
        ssize_t bytes_read = read(client_socket, buffer, min(remaining, (int)BUF_SIZE));
        if (bytes_read <= 0)
            break;
        body.append(buffer, bytes_read);
        remaining -= bytes_read;
    }
    return body;
}

// Sends the given response string to the client.
void sendResponse(int client_socket, const string &response) {
    ssize_t total_sent = 0;
    ssize_t to_send = response.size();
    const char *data = response.c_str();
    while (total_sent < to_send) {
        ssize_t sent = send(client_socket, data + total_sent, to_send - total_sent, 0);
        if (sent < 0) {
            perror("send");
            break;
        }
        total_sent += sent;
    }
}

// Reads the content of "csr.txt" from the current working directory.
string readCSR() {
    ifstream csrFile("csr.txt");
    if (!csrFile) {
        cerr << "Warning: Could not open jwt.json" << endl;
        return "";
    }
    ostringstream ss;
    ss << csrFile.rdbuf();
    string token = ss.str();
    if (!token.empty() && token.back() == '\n')
        token.pop_back();
    return token;
}

// Uses the base64 CLI tool to decode a base64-encoded string.
// It calls: echo '<encoded>' | base64 -d
string decodeBase64CLI(const string &encoded) {
    // Build the command.
    // Note: This simple command assumes that the encoded string does not contain single quotes.
    string command = "echo '" + encoded + "' | base64 -d";
    
    // Open a pipe to the command.
    FILE* pipe = popen(command.c_str(), "r");
    if (!pipe) {
        cerr << "Error: popen() failed for base64 command." << endl;
        return "";
    }
    
    char buffer[128];
    string decoded;
    while (fgets(buffer, sizeof(buffer), pipe) != nullptr) {
        decoded.append(buffer);
    }
    
    int returnCode = pclose(pipe);
    if (returnCode != 0) {
        cerr << "Error: base64 command returned non-zero exit code: " << returnCode << endl;
        return "";
    }
    
    return decoded;
}

int main() {
    // Create a TCP socket (IPv4).
    int server_fd = socket(AF_INET, SOCK_STREAM, 0);
    if (server_fd == -1) {
        perror("socket failed");
        exit(EXIT_FAILURE);
    }

    // Allow address reuse.
    int opt = 1;
    if (setsockopt(server_fd, SOL_SOCKET, SO_REUSEADDR, &opt, sizeof(opt)) < 0) {
        perror("setsockopt");
        exit(EXIT_FAILURE);
    }

    // Bind the socket to all interfaces on PORT.
    sockaddr_in address;
    memset(&address, 0, sizeof(address));
    address.sin_family = AF_INET;
    address.sin_addr.s_addr = INADDR_ANY;
    address.sin_port = htons(PORT);
    
    if (bind(server_fd, reinterpret_cast<sockaddr*>(&address), sizeof(address)) < 0) {
        perror("bind failed");
        exit(EXIT_FAILURE);
    }

    if (listen(server_fd, 3) < 0) {
        perror("listen");
        exit(EXIT_FAILURE);
    }

    cout << "Server listening on port " << PORT << endl;

    // --- Handle GET Request ---
    cout << "Waiting for GET request..." << endl;
    sockaddr_in client_addr;
    socklen_t client_addr_len = sizeof(client_addr);
    int client_socket = accept(server_fd, reinterpret_cast<sockaddr*>(&client_addr), &client_addr_len);
    if (client_socket < 0) {
        perror("accept");
        exit(EXIT_FAILURE);
    }
    
    // For GET, read until headers are received.
    string headers = readHeaders(client_socket);
    cout << "Received GET request:\n" << headers << endl;
    if (headers.compare(0, 3, "GET") != 0) {
        string error_response = "HTTP/1.1 400 Bad Request\r\n\r\n";
        sendResponse(client_socket, error_response);
        close(client_socket);
        close(server_fd);
        exit(EXIT_FAILURE);
    }

    // Build and send the GET response using the csr from csr.txt.
    string csr = readCSR();
    string get_response =
        "HTTP/1.1 200 OK\r\n"
        "Content-Type: application/json\r\n"
        "\r\n"
        "{\"csr\": \"" + csr + "\"}";
    sendResponse(client_socket, get_response);
    close(client_socket);
    cout << "GET handled." << endl;

    // --- Handle POST Request ---
    cout << "Waiting for POST request..." << endl;
    client_socket = accept(server_fd, reinterpret_cast<sockaddr*>(&client_addr), &client_addr_len);
    if (client_socket < 0) {
        perror("accept");
        close(server_fd);
        exit(EXIT_FAILURE);
    }
    
    // Read headers from the POST request.
    string postHeaders = readHeaders(client_socket);
    cout << "Received POST request headers:\n" << postHeaders << endl;
    if (postHeaders.compare(0, 4, "POST") != 0) {
        string error_response = "HTTP/1.1 400 Bad Request\r\n\r\n";
        sendResponse(client_socket, error_response);
        close(client_socket);
        close(server_fd);
        exit(EXIT_FAILURE);
    }
    
    // Determine the expected length of the body.
    int contentLength = getContentLength(postHeaders);
    if (contentLength < 0) {
        cerr << "Could not find Content-Length header." << endl;
        close(client_socket);
        close(server_fd);
        exit(EXIT_FAILURE);
    }
    
    // Extract any part of the body that was already read.
    size_t headerEndPos = postHeaders.find("\r\n\r\n");
    string body;
    if (headerEndPos != string::npos) {
        body = postHeaders.substr(headerEndPos + 4);
    }
    
    // Read the rest of the body if it hasn't been fully read yet.
    int alreadyRead = body.size();
    if (alreadyRead < contentLength) {
        body += readBody(client_socket, contentLength - alreadyRead);
    }
    
    cout << "Full POST body length: " << body.size() << " bytes" << endl;
    
    // Parse the JSON body.
    try {
        json j = json::parse(body);
        if (j.contains("cert") && j["cert"].is_string()) {
            string cert_base64 = j["cert"];
            // Decode using the base64 CLI.
            string decoded_cert = decodeBase64CLI(cert_base64);
            // Write the decoded certificate to fullchain.pem.
            ofstream outfile("fullchain.pem", ios::binary);
            if (!outfile) {
                cerr << "Error: Could not open fullchain.pem for writing." << endl;
            } else {
                outfile << decoded_cert;
                outfile.close();
                cout << "Saved decoded certificate to fullchain.pem" << endl;
            }
        } else {
            cerr << "JSON does not contain a valid 'cert' key." << endl;
        }

        // Check for 'reference' in the JSON body.
        if (j.contains("reference") && j["reference"].is_string()) {
            string reference_base64 = j["reference"];
            // Decode using the base64 CLI.
            string decoded_reference = decodeBase64CLI(reference_base64);
            // Write the decoded content to reference.json.
            ofstream outfile("reference.json", ios::binary);
            if (!outfile) {
                cerr << "Error: Could not open reference.json for writing." << endl;
            } else {
                outfile << decoded_reference;
                outfile.close();
                cout << "Saved decoded reference to reference.json" << endl;
            }
        } else {
            cerr << "JSON does not contain a valid 'reference' key." << endl;
        }
    } catch (json::parse_error &e) {
        cerr << "JSON parse error: " << e.what() << endl;
    } catch (exception &e) {
        cerr << "Error: " << e.what() << endl;
    }

    // Respond with 204 No Content.
    string post_response = "HTTP/1.1 204 No Content\r\n\r\n";
    sendResponse(client_socket, post_response);
    close(client_socket);
    cout << "POST handled." << endl;

    close(server_fd);
    cout << "Server done." << endl;
    return 0;
}
