BUILD_DIR=build
BIN_NAME=boot

build:
	cd source && go $(BUILD_DIR) -o ../$(BUILD_DIR)/$(BIN_NAME) .

clean:
	rm -rf $(BUILD_DIR)
