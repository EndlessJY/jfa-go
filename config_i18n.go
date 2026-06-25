package main

import (
	"encoding/json"

	"github.com/gin-gonic/gin"
	"github.com/hrfee/jfa-go/common"
)

type configI18nText struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type configI18nFile struct {
	Groups   map[string]configI18nText               `json:"groups"`
	Sections map[string]configI18nText               `json:"sections"`
	Settings map[string]map[string]configI18nText    `json:"settings"`
	Options  map[string]map[string]map[string]string `json:"options"`
}

func (app *appContext) localizedConfig(gc *gin.Context, conf common.Config) common.Config {
	lang := app.getLang(gc, AdminPage, app.storage.lang.chosenAdminLang)
	if lang != "zh-hans" && lang != "zh-hant" {
		return conf
	}

	translations, ok := loadConfigI18n(lang)
	if !ok {
		return conf
	}

	localized := cloneConfig(conf)
	for i := range localized.Groups {
		group := &localized.Groups[i]
		if text, ok := translations.Groups[group.Group]; ok {
			applyConfigI18nText(&group.Name, &group.Description, text)
		}
	}

	for i := range localized.Sections {
		section := &localized.Sections[i]
		if text, ok := translations.Sections[section.Section]; ok {
			applyConfigI18nText(&section.Meta.Name, &section.Meta.Description, text)
		}
		for j := range section.Settings {
			setting := &section.Settings[j]
			if sectionTranslations, ok := translations.Settings[section.Section]; ok {
				if text, ok := sectionTranslations[setting.Setting]; ok {
					applyConfigI18nText(&setting.Name, &setting.Description, text)
				}
			}
			if sectionOptions, ok := translations.Options[section.Section]; ok {
				if settingOptions, ok := sectionOptions[setting.Setting]; ok {
					for k := range setting.Options {
						if label, ok := settingOptions[setting.Options[k][0]]; ok {
							setting.Options[k][1] = label
						}
					}
				}
			}
		}
	}

	return localized
}

func loadConfigI18n(lang string) (configI18nFile, bool) {
	var translations configI18nFile
	data, err := langFS.ReadFile(FSJoin("config", lang+".json"))
	if err != nil {
		return translations, false
	}
	if err := json.Unmarshal(data, &translations); err != nil {
		return translations, false
	}
	return translations, true
}

func applyConfigI18nText(name *string, description *string, text configI18nText) {
	if text.Name != "" {
		*name = text.Name
	}
	if text.Description != "" {
		*description = text.Description
	}
}

func cloneConfig(conf common.Config) common.Config {
	out := conf
	out.Order = append([]common.Member(nil), conf.Order...)
	out.Groups = make([]common.Group, len(conf.Groups))
	for i, group := range conf.Groups {
		out.Groups[i] = group
		out.Groups[i].Members = append([]common.Member(nil), group.Members...)
	}
	out.Sections = make([]common.Section, len(conf.Sections))
	for i, section := range conf.Sections {
		out.Sections[i] = section
		out.Sections[i].Meta.Aliases = append([]string(nil), section.Meta.Aliases...)
		out.Sections[i].Settings = make([]common.Setting, len(section.Settings))
		for j, setting := range section.Settings {
			out.Sections[i].Settings[j] = setting
			out.Sections[i].Settings[j].Options = append([]common.Option(nil), setting.Options...)
			out.Sections[i].Settings[j].Aliases = append([]string(nil), setting.Aliases...)
		}
	}
	return out
}
