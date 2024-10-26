-- Save volume changes immediately to the config file
local config_file = os.getenv("HOME") .. "/.config/tera/tera_config.conf"

function save_volume()
    local volume = mp.get_property("volume")
    os.execute("mkdir -p " .. os.getenv("HOME") .. "/.config/tera")
    os.execute("echo volume=" .. volume .. " > " .. config_file)
end

-- Register the function to listen for volume change events
mp.observe_property("volume", "native", save_volume)
