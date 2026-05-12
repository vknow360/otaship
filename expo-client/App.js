import { StatusBar } from "expo-status-bar";
import { Alert, Image, StyleSheet, Text, View } from "react-native";
import * as Updates from "expo-updates";
import { useEffect } from "react";

export default function App() {
    useEffect(() => {
        console.log("=== UPDATES MODULE INFO ===");
        console.log("Is embedded launch:", Updates.isEmbeddedLaunch);
        console.log("Is emergency launch:", Updates.isEmergencyLaunch);
        console.log("Channel:", Updates.channel);
        console.log("Runtime version:", Updates.runtimeVersion);
        console.log("Update ID:", Updates.updateId);
        console.log("Updates enabled:", Updates.isEnabled);

        if (Updates.manifest) {
            console.log("Manifest detected:");
            console.log("- ID:", Updates.manifest.id);
            console.log("- Runtime Version:", Updates.manifest.runtimeVersion);
            console.log("- Created At:", Updates.manifest.createdAt);
        } else {
            console.log("No manifest currently loaded.");
        }

        if (!Updates.isEnabled) {
            console.log("Updates are DISABLED!");
            return;
        }

        async function checkUpdates() {
            try {
                console.log("Starting update check...");
                const update = await Updates.checkForUpdateAsync();
                console.log("Update check complete:", update);
                if (update.isAvailable) {
                    Alert.alert(
                        "New update available!",
                        "A new version of the app is available. Do you want to update now?",
                        [
                            { text: "Cancel", style: "cancel" },
                            {
                                text: "Update", onPress: () => {
                                    Updates.fetchUpdateAsync().then(() => {
                                        Updates.reloadAsync();
                                    })
                                }
                            },
                        ],
                        { cancelable: false }
                    )
                }
            } catch (e) {
                console.log("Update check failed!");
                console.log("Error message:", e.message);
                console.log("Error code:", e.code);
                if (e.stack) console.log("Stack trace:", e.stack);
            }
        }

        setTimeout(checkUpdates, 3000);
    }, []);
    return (
        <View style={styles.container}>
            <Text style={{ color: "black" }}>This is test app</Text>

            <Image source={require("./assets/favicon.png")} />
            <StatusBar style="auto" backgroundColor="black" />
        </View>
    );
}

const styles = StyleSheet.create({
    container: {
        flex: 1,
        backgroundColor: "#fff",
        alignItems: "center",
        justifyContent: "center",
    },
});
