package helpers

const cdxBashEnvTemplate = `#!/usr/bin/env bash

function show_help() {
    echo "\nUsage: mri [command]\n"
    echo "Available commands:\n"
    echo "  help                    📖 Show this help message"
    echo "  handle_apt_unmet        🔧 Fix broken packages in APT"
    echo "  installer               📦 Interactive installer (APT/DPKG)"
    echo "  la                      📂 Enhanced file listing"
    echo "  get_public_ip           🌍 Get public IP"
    echo "  xcp                     📋 Copy output to clipboard (XClip)"
    echo "  reloadrc                🔄 Reload shell configuration"
    echo "  cls                     🧹 Clear the screen"
    echo "  wtf_is_that             🔍 Find information about a command/program"
    echo "  upgrade-fix             🔄 Update and clean the system (enhanced)"
    echo "  safe_upgrade            🚧 Ask before updating the system"
    echo "\nExample: mri la"
}

function mri() {
    case "$1" in
        help) show_help ;;
        handle_apt_unmet) mri_handle_apt_unmet ;;
        installer) mri_installer ;;
        la) la ;;
        get_public_ip) get_public_ip ;;
        xcp) shift; xcp "$@" ;;
        reloadrc) reloadrc ;;
        cls) cls ;;
        wtf_is_that) shift; wtf_is_that "$@" ;;
        upgrade-fix) upgrade-fix ;;
        safe_upgrade) safe_upgrade ;;
        *) echo "❌ Unknown command. Use 'mri help'." ;;
    esac
}

## 🚀 Auto-loading Node.js (via NVM)
if [[ -d "$HOME/.nvm" ]]; then
  export NVM_DIR="$HOME/.nvm"
  [ -s "$NVM_DIR/nvm.sh" ] && . "$NVM_DIR/nvm.sh"
  [ -s "$NVM_DIR/bash_completion" ] && . "$NVM_DIR/bash_completion"
fi

## 🚀 Load Mori Logger, if available
if [[ ! $(command -v mri_log) && -f "${HOME}/.mori_logger_env" ]]; then
	. "${HOME}/.mori_logger_env" || true
fi

## 🚀 Fix broken packages in APT
function mri_handle_apt_unmet() {
	if [ "$1" -eq 100 ]; then
		sudo apt --fix-broken install
	fi
}

## 🚀 Interactive installer (APT/DPKG)
function mri_installer() {
	if [[ -z "$1" ]]; then
		mri_log error "❌ You need to specify a package to install!"
		return 1
	fi

	if [ ! -f "${1}" ]; then
		mri_log warn "Calm down man! One by one, otherwise the PC will fry..."
		return 1;
	fi

	if command -v dialog; then
		HEIGHT=15
		WIDTH=40
		CHOICE_HEIGHT=4
		BACKTITLE="Installer Selection"
		TITLE="Choose the Right Installer"
		MENU="Choose one of the following options:"

		OPTIONS=(1 "Install with APT"
			     2 "Install with DPKG")

		CHOICE=$(dialog --clear \
			            --backtitle "$BACKTITLE" \
			            --title "$TITLE" \
			            --menu "$MENU" \
			            $HEIGHT $WIDTH $CHOICE_HEIGHT \
			            "${OPTIONS[@]}" \
			            2>&1 >/dev/tty)

		clear
		case $CHOICE in
			1)
			    mri_log info "Installing with APT..."
			    sudo apt install "${1}" || cdx_handle_apt_unmet $?
			    ;;
			2)
			    mri_log info "Installing with DPKG..."
			    sudo dpkg -i "${1}"
			    ;;
		esac
	else
		HEIGHT=15
		WIDTH=40
		CHOICE_HEIGHT=4
		BACKTITLE="Installer Selection"
		TITLE="Choose the Right Installer"
		MENU="Choose one of the following options:"

		OPTIONS=(1 "Install with APT"
			     2 "Install with DPKG")

		CHOICE=$(whiptail --clear \
			              --backtitle "$BACKTITLE" \
			              --title "$TITLE" \
			              --menu "$MENU" \
			              $HEIGHT $WIDTH $CHOICE_HEIGHT \
			              "${OPTIONS[@]}" \
			              3>&1 1>&2 2>&3)

		clear
		case $CHOICE in
			1)
			    mri_log info "Installing with APT..."
			    sudo apt install "${1}" || cdx_handle_apt_unmet $?
			    ;;
			2)
			    mri_log info "Installing with DPKG..."
			    sudo dpkg -i "${1}"
			    ;;
		esac
	fi
}

## 🚀 Enhanced file listing
function la() {
	local dir="${1:-.}"
	if [[ ! -d "$dir" ]]; then
		mri_log error "❌ '$dir' is not a valid directory!"
		return 1
	fi
	ls -lAh --color=always --group-directories-first "$dir"
}

## 🚀 Get public IP
function get_public_ip() {
	curl -s https://api.ipify.org || {
		mri_log error "❌ Failed to get public IP."
		return 1
	}
}

## 🚀 Copy output to clipboard (XClip)
function xcp() {
	if command -v xclip &>/dev/null; then
		echo -e "$*" | xclip -selection clipboard
	else
		mri_log error "❌ xclip is not installed."
		return 1
	fi
}

## 🚀 Reload shell configuration
function reloadrc() {
	local shell_rc="$HOME/.${SHELL##*/}rc"
	if [[ -f "$shell_rc" ]]; then
		. "$shell_rc"
		mri_log success "✅ $USER Configurations reloaded!"
	else
		mri_log error "❌ Configuration file not found: $shell_rc"
		return 1
	fi
}

## 🚀 Clear the screen
function cls() {
	tput reset || clear
}

## 🚀 Find information about a command/program
function wtf_is_that() {
	local cmd="$1"
	local response=""

	for tool in whatis command -V type which whereis apropos; do
		response=$($tool "$cmd" 2>/dev/null) && break
	done

	if [[ -n "$response" ]]; then
		mri_log info "🔍 $response"
	else
		mri_log error "❌ No information found for '$cmd'."
		return 1
	fi
}

## 🚀 Update and clean the system (enhanced)
function upgrade-fix() {
	mri_log info "🔄 Updating packages..."
	sudo apt-get update -y && sudo apt full-upgrade -y --fix-broken
	mri_log info "🧹 Cleaning unnecessary packages..."
	sudo apt autoremove -y && sudo apt autoclean -y && sudo apt purge -y
	mri_log success "✅ Update completed!"
}

## 🔹 Ask before updating the system
function safe_upgrade() {
    if bash $HOME/.mori_yes_no_question "Do you want to update the system now?" "n" 5; then
        upgrade-fix
    else
        mri_log warn "🚧 Update canceled by the user."
    fi
}

`
