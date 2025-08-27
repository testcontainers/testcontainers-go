for i in $(seq 1 ${ACTIVATOR_RETRY}); do
					echo "nebula" | nebula-console -addr graphd0 -port 9669 -u root -e 'ADD HOSTS "storaged0":9779' 1>/dev/null 2>/dev/null
					if [ $? -eq 0 ]; then
						echo "✔️ Storage activated successfully."
						break
					else
						output=$(echo "nebula" | nebula-console -addr graphd0 -port 9669 -u root -e 'ADD HOSTS "storaged0":9779' 2>&1)
						if echo "$output" | grep -q "Existed"; then
							echo "✔️ Storage already activated, Exiting..."
							break
						fi
					fi
					if [ $i -lt ${ACTIVATOR_RETRY} ]; then
						echo "⏳ Attempting to activate storaged, attempt $i/${ACTIVATOR_RETRY}... It's normal to take some attempts before storaged is ready. Please wait."
					else
						echo "❌ Failed to activate storaged after ${ACTIVATOR_RETRY} attempts. Please check MetaD, StorageD logs."
						echo "ℹ️ Error during storage activation:"
						echo "=============================================================="
						echo "$output"
						echo "=============================================================="
						exit 1
					fi
					sleep 5
				done && tail -f /dev/null