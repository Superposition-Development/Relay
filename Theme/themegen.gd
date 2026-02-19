@tool
extends ProgrammaticTheme

func setup():
	set_save_path("res://Theme/RelayBase/RelayBase.tres")

func define_theme():
	var panel_base = stylebox_flat({
		"anti_aliasing": true,
		"anti_aliasing_size": 1.0,
		"bg_color": Color(0.089, 0.089, 0.089, 1.0),
		"corner_radius_bottom_left": 5,
		"corner_radius_bottom_right": 5,
		"corner_radius_top_left": 5,
		"corner_radius_top_right": 5,
	})
	
	var panel_surface_1 = inherit(panel_base, {
		"bg_color": Color(0.137, 0.137, 0.137, 1.0),
	})

	var panel_bg = inherit(panel_base, {
		"bg_color": Color(0.038, 0.038, 0.038, 1.0),
		"corner_radius_bottom_left": 0,
		"corner_radius_bottom_right": 0,
		"corner_radius_top_left": 0,
		"corner_radius_top_right": 0,
	})
	
	var text_edit_base = stylebox_flat({
		"bg_color": Color(0.0, 0.0, 0.0, 0.0)
	})
	
	define_style("MarginContainer", {
		"margin_left": 10,
		"margin_right": 10,
		"margin_top": 10,
		"margin_bottom": 10,
	})
	
	define_variant_style("MarginContainerSmall", "MarginContainer", {
		"margin_left": 5,
		"margin_right": 5,
		"margin_top": 5,
		"margin_bottom": 5,
	})
	
	define_variant_style("MarginContainerMicro", "MarginContainer", {
		"margin_left": 2,
		"margin_right": 2,
		"margin_top": 2,
		"margin_bottom": 2,
	})
	
	define_style("PanelContainer", {
		"panel": panel_base
	})
	
	define_variant_style("BG", "PanelContainer", {
		"panel": panel_bg
	})
	
	define_variant_style("PanelSurface1", "PanelContainer", {
		"panel": panel_surface_1
	})
	
	define_style("LineEdit", {
		"normal": text_edit_base,
		"read_only": text_edit_base,
		"focus": text_edit_base,
	})
	
	
	
